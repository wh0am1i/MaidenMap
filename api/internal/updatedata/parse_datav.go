package updatedata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// DefaultDataVBaseURL is Alibaba DataV's public GeoJSON endpoint for
// Chinese administrative boundaries. Every adcode/_full.json returns the
// direct children of that admin node.
const DefaultDataVBaseURL = "https://geo.datav.aliyun.com/areas_v3/bound"

// dataVMaxBodyBytes caps a single DataV response. The country-level file
// is the largest real payload at roughly 30 MB; 100 MB leaves headroom and
// bounds the OOM surface if an endpoint starts streaming arbitrary data.
// Overridable in tests.
var dataVMaxBodyBytes int64 = 100 << 20 // 100 MiB

// dataVRequestTimeout bounds any single _full.json fetch. Independent of
// the shared 30-minute http.Client timeout so a slow drill at the tail
// doesn't let early fetches run indefinitely.
var dataVRequestTimeout = 60 * time.Second

// dataVMaxFailureRate is the per-level fail-tolerance for level-2 and level-3
// drills. Beyond this, FetchDataVChina returns an error so the caller can
// leave the previous datav.geojson in place rather than atomically replacing
// it with a half-populated file.
const dataVMaxFailureRate = 0.20

// flexInt accepts either a JSON number or a JSON string. DataV's level-1
// response includes a marker feature ("100000_JD" for the Nine-Dash Line
// polyline) whose `adcode` is a non-numeric string — decoding that via a
// plain int field blows up the whole FeatureCollection. Non-numeric strings
// decode to 0 here and downstream code already skips features with adcode 0.
type flexInt int

func (f *flexInt) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	if b[0] == '"' {
		// Strip quotes; accept only strings that parse cleanly as ints —
		// "100000_JD" must NOT decode to 100000 since that would collide
		// with the real country root.
		s := string(b[1 : len(b)-1])
		if n, err := strconv.Atoi(s); err == nil {
			*f = flexInt(n)
		}
		return nil
	}
	n := 0
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*f = flexInt(n)
	return nil
}

// DataVProperties is the subset of GeoJSON feature properties we care about.
type DataVProperties struct {
	ADCode      flexInt `json:"adcode"`
	Name        string  `json:"name"`
	Level       string  `json:"level"` // country / province / city / district
	ChildrenNum int     `json:"childrenNum"`
	Parent      struct {
		ADCode flexInt `json:"adcode"`
	} `json:"parent"`
}

// dataVFeatureRaw mirrors a single DataV feature with the full geometry kept
// as raw JSON so we can pass it through to the emitted file unchanged.
type dataVFeatureRaw struct {
	Type       string          `json:"type"`
	Properties DataVProperties `json:"properties"`
	Geometry   json.RawMessage `json:"geometry"`
}

type dataVFeatureCollection struct {
	Type     string            `json:"type"`
	Features []dataVFeatureRaw `json:"features"`
}

// DataVNode is the output shape after assembly: an admin area with a parent
// link and geometry. Callers can marshal this into a slimmed GeoJSON file.
type DataVNode struct {
	ADCode   int             `json:"adcode"`
	Name     string          `json:"name"`
	Level    string          `json:"level"`
	ParentAD int             `json:"parent"`
	Geometry json.RawMessage `json:"geometry"`
}

// FetchDataVChina walks DataV's admin tree from the country root down to
// districts (where available), returning a flat list of admin nodes with
// their polygons. Network-heavy — roughly one HTTP call per province plus
// one per mainland city (~370 total). Fetches run in parallel.
//
// Returns an error — and no nodes — if the per-level failure rate exceeds
// dataVMaxFailureRate. That lets the caller leave the existing on-disk
// datav.geojson untouched rather than atomically replacing it with a
// half-populated file when the upstream is flaky.
func FetchDataVChina(baseURL string, concurrency int) ([]DataVNode, error) {
	if baseURL == "" {
		baseURL = DefaultDataVBaseURL
	}
	if concurrency < 1 {
		concurrency = 8
	}

	ctx := context.Background()

	// Level 1: provinces under country (100000).
	root, err := fetchDataV(ctx, baseURL, 100000)
	if err != nil {
		return nil, fmt.Errorf("fetch country: %w", err)
	}
	slog.Info("datav country fetched", "provinces", len(root.Features))

	out := make([]DataVNode, 0, 4096)
	for _, f := range root.Features {
		if int(f.Properties.ADCode) == 0 {
			continue
		}
		out = append(out, toNode(f, 100000))
	}

	// Level 2: children of each province (cities or, for HK/MO, districts).
	type level2Result struct {
		parent int
		fc     *dataVFeatureCollection
		err    error
	}
	level2Ch := make(chan level2Result, len(root.Features))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var level2Issued int
	for _, p := range root.Features {
		ad := int(p.Properties.ADCode)
		if ad == 0 || p.Properties.ChildrenNum == 0 {
			continue
		}
		level2Issued++
		wg.Add(1)
		sem <- struct{}{}
		go func(adcode int) {
			defer wg.Done()
			defer func() { <-sem }()
			fc, err := fetchDataV(ctx, baseURL, adcode)
			level2Ch <- level2Result{parent: adcode, fc: fc, err: err}
		}(ad)
	}
	go func() { wg.Wait(); close(level2Ch) }()

	// Collect level-2 results and queue level-3 fetches for city-level nodes.
	type level3Job struct {
		parent int
	}
	var level3Jobs []level3Job
	var level2Fails atomic.Int32
	for r := range level2Ch {
		if r.err != nil {
			// TW 710000 has no _full; we still treat it as a fail for rate
			// purposes — dataVMaxFailureRate is set loose enough (20 %) that
			// the handful of expected no-drill provinces don't trip it, but
			// a systemic outage will.
			level2Fails.Add(1)
			slog.Warn("datav level2 fetch failed", "parent", r.parent, "err", r.err)
			continue
		}
		for _, f := range r.fc.Features {
			if int(f.Properties.ADCode) == 0 {
				continue
			}
			out = append(out, toNode(f, r.parent))
			if f.Properties.Level == "city" && f.Properties.ChildrenNum > 0 {
				level3Jobs = append(level3Jobs, level3Job{parent: int(f.Properties.ADCode)})
			}
		}
	}
	if err := checkFailureRate("level2", int(level2Fails.Load()), level2Issued); err != nil {
		return nil, err
	}
	slog.Info("datav level2 complete", "nodes_so_far", len(out), "cities_to_drill", len(level3Jobs), "fails", level2Fails.Load(), "issued", level2Issued)

	// Level 3: districts under each city. Fan out with the same worker limit.
	type level3Result struct {
		parent int
		fc     *dataVFeatureCollection
		err    error
	}
	level3Ch := make(chan level3Result, len(level3Jobs))
	var wg3 sync.WaitGroup
	for _, job := range level3Jobs {
		wg3.Add(1)
		sem <- struct{}{}
		go func(j level3Job) {
			defer wg3.Done()
			defer func() { <-sem }()
			fc, err := fetchDataV(ctx, baseURL, j.parent)
			level3Ch <- level3Result{parent: j.parent, fc: fc, err: err}
		}(job)
	}
	go func() { wg3.Wait(); close(level3Ch) }()

	var level3Fails atomic.Int32
	for r := range level3Ch {
		if r.err != nil {
			level3Fails.Add(1)
			slog.Warn("datav level3 fetch failed", "parent", r.parent, "err", r.err)
			continue
		}
		for _, f := range r.fc.Features {
			if int(f.Properties.ADCode) == 0 {
				continue
			}
			out = append(out, toNode(f, r.parent))
		}
	}
	if err := checkFailureRate("level3", int(level3Fails.Load()), len(level3Jobs)); err != nil {
		return nil, err
	}

	slog.Info("datav drill complete", "total_nodes", len(out), "level3_fails", level3Fails.Load(), "level3_issued", len(level3Jobs))
	return out, nil
}

func checkFailureRate(level string, fails, issued int) error {
	if issued == 0 {
		return nil
	}
	rate := float64(fails) / float64(issued)
	if rate > dataVMaxFailureRate {
		return fmt.Errorf("datav %s failure rate %.0f%% exceeds %.0f%% (fails=%d issued=%d); previous datav.geojson left in place", level, rate*100, dataVMaxFailureRate*100, fails, issued)
	}
	return nil
}

func fetchDataV(ctx context.Context, baseURL string, adcode int) (*dataVFeatureCollection, error) {
	url := fmt.Sprintf("%s/%d_full.json", baseURL, adcode)
	ctx, cancel := context.WithTimeout(ctx, dataVRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, dataVMaxBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > dataVMaxBodyBytes {
		return nil, fmt.Errorf("GET %s: body exceeds cap (%d bytes)", url, dataVMaxBodyBytes)
	}
	var fc dataVFeatureCollection
	if err := json.Unmarshal(body, &fc); err != nil {
		return nil, fmt.Errorf("decode %s: %w", url, err)
	}
	return &fc, nil
}

func toNode(f dataVFeatureRaw, parent int) DataVNode {
	if f.Properties.Parent.ADCode != 0 {
		parent = int(f.Properties.Parent.ADCode)
	}
	return DataVNode{
		ADCode:   int(f.Properties.ADCode),
		Name:     f.Properties.Name,
		Level:    f.Properties.Level,
		ParentAD: parent,
		Geometry: f.Geometry,
	}
}

// EncodeDataVNodes writes the nodes as a slim FeatureCollection to w.
func EncodeDataVNodes(w io.Writer, nodes []DataVNode) error {
	type outFeature struct {
		Type       string          `json:"type"`
		Properties map[string]any  `json:"properties"`
		Geometry   json.RawMessage `json:"geometry"`
	}
	type outFC struct {
		Type     string       `json:"type"`
		Features []outFeature `json:"features"`
	}

	fc := outFC{Type: "FeatureCollection", Features: make([]outFeature, 0, len(nodes))}
	for _, n := range nodes {
		props := map[string]any{
			"adcode": n.ADCode,
			"name":   n.Name,
			"level":  n.Level,
			"parent": n.ParentAD,
		}
		fc.Features = append(fc.Features, outFeature{
			Type:       "Feature",
			Properties: props,
			Geometry:   n.Geometry,
		})
	}
	enc := json.NewEncoder(w)
	return enc.Encode(fc)
}

