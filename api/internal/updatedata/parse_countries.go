package updatedata

import (
	"github.com/paulmach/orb/geojson"
)

// ParseNaturalEarthCountries reads the Natural Earth country GeoJSON (properties
// ISO_A2_EH + NAME_EN) and returns a FeatureCollection normalized to lowercase keys.
func ParseNaturalEarthCountries(raw []byte) (*geojson.FeatureCollection, error) {
	fc, err := geojson.UnmarshalFeatureCollection(raw)
	if err != nil {
		return nil, err
	}
	for _, f := range fc.Features {
		iso := ""
		name := ""
		if v, ok := f.Properties["ISO_A2_EH"].(string); ok {
			iso = v
		} else if v, ok := f.Properties["iso_a2"].(string); ok {
			iso = v
		}
		if v, ok := f.Properties["NAME_EN"].(string); ok {
			name = v
		} else if v, ok := f.Properties["name_en"].(string); ok {
			name = v
		}
		f.Properties = map[string]any{"iso_a2": iso, "name_en": name}
	}
	return fc, nil
}
