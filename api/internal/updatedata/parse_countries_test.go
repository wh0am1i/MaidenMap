package updatedata

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNaturalEarthCountries(t *testing.T) {
	b, err := os.ReadFile("testdata/ne_countries.geojson")
	require.NoError(t, err)

	out, err := ParseNaturalEarthCountries(b)
	require.NoError(t, err)
	require.Len(t, out.Features, 4)

	byISO := map[string]map[string]any{}
	for _, f := range out.Features {
		byISO[f.Properties["iso_a2"].(string)] = f.Properties
	}

	assert.Equal(t, "Germany", byISO["DE"]["name_en"])
	assert.Equal(t, "德国", byISO["DE"]["name_zh"])
	assert.Equal(t, "日本", byISO["JP"]["name_zh"])

	// Mainland-friendly overrides: 香港 → 中国香港, 台湾 → 中国台湾.
	assert.Equal(t, "中国香港", byISO["HK"]["name_zh"])
	assert.Equal(t, "中国台湾", byISO["TW"]["name_zh"])
}
