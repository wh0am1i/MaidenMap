package updatedata

import (
	"encoding/json"
	"io"
)

// CNAdminNode is the in-memory shape of one Chinese administrative-division
// polygon flowing from parser → encoder. The output GeoJSON's properties
// follow the same names (adcode / name / level / parent), so the runtime
// CN-admin index can ingest the file directly.
type CNAdminNode struct {
	ADCode   int             `json:"adcode"`
	Name     string          `json:"name"`
	Level    string          `json:"level"`
	ParentAD int             `json:"parent"`
	Geometry json.RawMessage `json:"geometry"`
}

// EncodeCNAdmin writes nodes as a slim FeatureCollection to w. One feature
// per node, properties carry adcode / name / level / parent verbatim.
func EncodeCNAdmin(w io.Writer, nodes []CNAdminNode) error {
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
		fc.Features = append(fc.Features, outFeature{
			Type: "Feature",
			Properties: map[string]any{
				"adcode": n.ADCode,
				"name":   n.Name,
				"level":  n.Level,
				"parent": n.ParentAD,
			},
			Geometry: n.Geometry,
		})
	}
	return json.NewEncoder(w).Encode(fc)
}
