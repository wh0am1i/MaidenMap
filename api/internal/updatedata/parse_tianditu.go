package updatedata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Tianditu's National Geographic Information Public Service Platform ships
// the official "China standard map" administrative-division vector data under
// audit number GS(2024)0650. It comes in three files (province / city /
// county) downloaded from https://cloudcenter.tianditu.gov.cn/administrativeDivision/.
//
// Each feature carries:
//   - gb:   GB/T 2260 行政区划代码 prefixed with "156" (China's ISO 3166-1
//           numeric code), e.g. "156540423" = 墨脱县.
//   - name: Chinese name only (no English).
//
// The level information is implicit in which file the feature lives in, not
// in any property — so the loader threads the level through explicitly.

const (
	// countryAD is the synthetic root adcode used as parent for all provinces.
	// Matches what the prior DataV pipeline emitted, so the CN-admin index's
	// hierarchy walk is unchanged.
	countryAD = 100000

	tianditProvinceFile = "中国_省.geojson"
	tianditCityFile     = "中国_市.geojson"
	tianditCountyFile   = "中国_县.geojson"
)

type tianditFeature struct {
	Type       string `json:"type"`
	Properties struct {
		GB   string `json:"gb"`
		Name string `json:"name"`
	} `json:"properties"`
	Geometry json.RawMessage `json:"geometry"`
}

type tianditFeatureCollection struct {
	Type     string           `json:"type"`
	Features []tianditFeature `json:"features"`
}

// patchFeature mirrors a feature in data/disputed_patch.geojson. The patch
// file carries the full CNAdminNode shape in properties (adcode + level +
// parent), since the patches don't follow GB/T 2260 — they're synthetic
// polygons we draw ourselves to cover gaps the official county data leaves
// in disputed regions (e.g. eastern 藏南 between 墨脱县 and 察隅县).
type patchFeature struct {
	Type       string `json:"type"`
	Properties struct {
		ADCode int    `json:"adcode"`
		Name   string `json:"name"`
		Level  string `json:"level"`
		Parent int    `json:"parent"`
	} `json:"properties"`
	Geometry json.RawMessage `json:"geometry"`
}

type patchFeatureCollection struct {
	Type     string         `json:"type"`
	Features []patchFeature `json:"features"`
}

// LoadTianditu reads the three Tianditu administrative-division files from
// dir, plus an optional disputed-territory patch file at patchPath. Returns
// the flat node list ready for EncodeCNAdmin.
//
// patchPath may be "" to skip the patch; the file not existing is also OK
// (the pipeline still works, just without the patch polygons).
func LoadTianditu(dir, patchPath string) ([]CNAdminNode, error) {
	out := make([]CNAdminNode, 0, 4096)
	steps := []struct{ file, level string }{
		{tianditProvinceFile, "province"},
		{tianditCityFile, "city"},
		{tianditCountyFile, "district"},
	}
	for _, s := range steps {
		nodes, err := loadTianditFile(filepath.Join(dir, s.file), s.level)
		if err != nil {
			return nil, err
		}
		out = append(out, nodes...)
	}

	if patchPath != "" {
		patches, err := loadPatch(patchPath)
		if err != nil {
			return nil, err
		}
		out = append(out, patches...)
	}
	repairDistrictParents(out)
	return out, nil
}

// repairDistrictParents fixes districts whose derived parent doesn't exist in
// the index. The default rule (drop last two digits) assumes a 地级市 sits
// between district and province; for HK / MO / 县级市-directly-under-province
// there's no such middle tier, so the derived code lands in nothing. We
// re-target those to the province (drop last four digits).
//
// Without this fix, the hierarchy walk would orphan e.g. 中西区 (156810101 →
// parent 810100 doesn't exist), leaving admin1 empty.
func repairDistrictParents(nodes []CNAdminNode) {
	known := make(map[int]bool, len(nodes))
	for _, n := range nodes {
		known[n.ADCode] = true
	}
	for i, n := range nodes {
		if n.Level != "district" {
			continue
		}
		if known[n.ParentAD] {
			continue
		}
		if p := (n.ADCode / 10000) * 10000; known[p] {
			nodes[i].ParentAD = p
		}
	}
}

func loadTianditFile(path, level string) ([]CNAdminNode, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var fc tianditFeatureCollection
	if err := json.Unmarshal(b, &fc); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	out := make([]CNAdminNode, 0, len(fc.Features))
	for _, f := range fc.Features {
		// The 省 file ships ~9 "境界线" (boundary line) features with empty
		// gb and non-polygon geometry — drop them; they're not lookup targets.
		if f.Properties.GB == "" {
			continue
		}
		ad, parent, ok := gbToAdcode(f.Properties.GB, level)
		if !ok {
			continue
		}
		// SAR / TW deduplication: 香港特别行政区 / 澳门特别行政区 / 台湾省
		// can appear in 省, 市, AND 县 files all with the same gb 156XX0000.
		// Only the 省 entry is meaningful — the duplicates would overwrite
		// it in byADCode (since map insertion is last-wins) and break the
		// hierarchy walk. The province polygon is identical anyway.
		if level != "province" && ad%10000 == 0 {
			continue
		}
		out = append(out, CNAdminNode{
			ADCode:   ad,
			Name:     f.Properties.Name,
			Level:    level,
			ParentAD: parent,
			Geometry: f.Geometry,
		})
	}
	return out, nil
}

func loadPatch(path string) ([]CNAdminNode, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read patch %s: %w", path, err)
	}
	var fc patchFeatureCollection
	if err := json.Unmarshal(b, &fc); err != nil {
		return nil, fmt.Errorf("decode patch %s: %w", path, err)
	}
	out := make([]CNAdminNode, 0, len(fc.Features))
	for _, f := range fc.Features {
		if f.Properties.ADCode == 0 || f.Properties.Level == "" {
			continue
		}
		out = append(out, CNAdminNode{
			ADCode:   f.Properties.ADCode,
			Name:     f.Properties.Name,
			Level:    f.Properties.Level,
			ParentAD: f.Properties.Parent,
			Geometry: f.Geometry,
		})
	}
	return out, nil
}

// gbToAdcode converts a Tianditu 9-digit gb string into a 6-digit adcode and
// derives the parent adcode by the level rule:
//
//   - province → parent = countryAD (100000)
//   - city     → parent = province part (last 4 digits zeroed)
//   - district → parent = city part (last 2 digits zeroed)
//
// District codes whose derived city doesn't exist (县级市 sitting directly
// under a province, e.g. 石河子市 / 659001) are handled by the index's
// hierarchyFor walk: an unresolved parent link just leaves the City field
// empty, the province lookup further up still works.
func gbToAdcode(gb, level string) (int, int, bool) {
	if len(gb) != 9 || gb[:3] != "156" {
		return 0, 0, false
	}
	ad, err := strconv.Atoi(gb[3:])
	if err != nil || ad == 0 {
		return 0, 0, false
	}
	switch level {
	case "province":
		return ad, countryAD, true
	case "city":
		return ad, (ad / 10000) * 10000, true
	case "district":
		// 县级市 directly under a province (中山市 442000, 东莞市 441900,
		// 儋州市 460400, 嘉峪关市 620200, etc.) sit in the 县 file but have
		// the prefecture slot zeroed (XXYY00 with YY != 00 and trailing 00).
		// For them the prefecture parent is the province itself, not a
		// phantom 442000 self-reference. Detect ad%100==0 ⇒ skip the
		// prefecture tier.
		if ad%100 == 0 {
			return ad, (ad / 10000) * 10000, true
		}
		return ad, (ad / 100) * 100, true
	}
	return 0, 0, false
}
