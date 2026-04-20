package updatedata

import (
	"github.com/paulmach/orb/geojson"
)

// countryNameZhOverrides lets us correct zh country names that Natural Earth
// gives verbatim (香港 / 澳门 / 台湾) into forms that are clearer in mainland
// context (中国香港 / 中国澳门 / 中国台湾). Keys are ISO 3166-1 alpha-2.
var countryNameZhOverrides = map[string]string{
	"HK": "中国香港",
	"MO": "中国澳门",
	"TW": "中国台湾",
}

// ParseNaturalEarthCountries reads the Natural Earth country GeoJSON (properties
// ISO_A2_EH + NAME_EN + NAME_ZH) and returns a FeatureCollection normalized to
// lowercase keys (iso_a2 / name_en / name_zh). Applies zh-name overrides for
// a small set of territories where the Natural Earth value is politically
// neutral but mainland users expect the longer form.
func ParseNaturalEarthCountries(raw []byte) (*geojson.FeatureCollection, error) {
	fc, err := geojson.UnmarshalFeatureCollection(raw)
	if err != nil {
		return nil, err
	}
	for _, f := range fc.Features {
		iso := pickString(f.Properties, "ISO_A2_EH", "iso_a2")
		nameEn := pickString(f.Properties, "NAME_EN", "name_en")
		nameZh := pickString(f.Properties, "NAME_ZH", "name_zh")
		if override, ok := countryNameZhOverrides[iso]; ok {
			nameZh = override
		}
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
