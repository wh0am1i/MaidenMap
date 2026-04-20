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
	require.Len(t, out.Features, 2)

	assert.Equal(t, "DE", out.Features[0].Properties["iso_a2"])
	assert.Equal(t, "Germany", out.Features[0].Properties["name_en"])
	assert.Equal(t, "德国", out.Features[0].Properties["name_zh"])

	assert.Equal(t, "JP", out.Features[1].Properties["iso_a2"])
	assert.Equal(t, "日本", out.Features[1].Properties["name_zh"])
}
