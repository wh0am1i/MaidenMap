package updatedata

import (
	"github.com/paulmach/orb/geojson"
)

// ParseNaturalEarthCountries reads the Natural Earth country GeoJSON (properties
// ISO_A2_EH + NAME_EN + NAME_ZH) and returns a FeatureCollection normalized to
// lowercase keys (iso_a2 / name_en / name_zh).
//
// HK / MO / TW are kept as their own features here — the handler-layer
// transform (see handler.applyChinaSARTransform) folds them into CN at
// response time, which keeps the raw data honest to Natural Earth while
// still giving a mainland-friendly display.
func ParseNaturalEarthCountries(raw []byte) (*geojson.FeatureCollection, error) {
	fc, err := geojson.UnmarshalFeatureCollection(raw)
	if err != nil {
		return nil, err
	}
	for _, f := range fc.Features {
		iso := pickString(f.Properties, "ISO_A2_EH", "iso_a2")
		nameEn := pickString(f.Properties, "NAME_EN", "name_en")
		nameZh := pickString(f.Properties, "NAME_ZH", "name_zh")
		f.Properties = map[string]any{
			"iso_a2":  iso,
			"name_en": nameEn,
			"name_zh": nameZh,
		}
	}
	return fc, nil
}

func pickString(p map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := p[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}
