package updatedata

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCitiesGeoNames(t *testing.T) {
	f, err := os.Open("testdata/cities_sample.txt")
	require.NoError(t, err)
	defer f.Close()

	cities, err := ParseCitiesGeoNames(f)
	require.NoError(t, err)
	require.Len(t, cities, 2)

	assert.Equal(t, "Berlin", cities[0].Name)
	assert.InDelta(t, 52.524, cities[0].Lat, 1e-2)
	assert.InDelta(t, 13.410, cities[0].Lon, 1e-2)
	assert.Equal(t, "DE", cities[0].CountryCode)
	assert.Equal(t, "16", cities[0].Admin1Code)
	assert.Equal(t, "11000", cities[0].Admin2Code)

	assert.Equal(t, "Tokyo", cities[1].Name)
	assert.Equal(t, "JP", cities[1].CountryCode)
}
