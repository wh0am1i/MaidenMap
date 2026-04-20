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

	entries, err := ParseCitiesGeoNames(f)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, uint32(2950159), entries[0].GeonameID)
	assert.Equal(t, "Berlin", entries[0].City.Name)
	assert.Empty(t, entries[0].City.NameZh)
	assert.InDelta(t, 52.524, entries[0].City.Lat, 1e-2)
	assert.InDelta(t, 13.410, entries[0].City.Lon, 1e-2)
	assert.Equal(t, "DE", entries[0].City.CountryCode)
	assert.Equal(t, "16", entries[0].City.Admin1Code)
	assert.Equal(t, "11000", entries[0].City.Admin2Code)

	assert.Equal(t, uint32(1850147), entries[1].GeonameID)
	assert.Equal(t, "Tokyo", entries[1].City.Name)
	assert.Equal(t, "JP", entries[1].City.CountryCode)
}
