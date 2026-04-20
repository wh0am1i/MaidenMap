package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wh0am1i/maidenmap/api/internal/geocode"
)

func TestLoadFromDir(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"countries.geojson", "admin_codes.json"} {
		src := filepath.Join("testdata", name)
		dst := filepath.Join(dir, name)
		b, err := os.ReadFile(src)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dst, b, 0o644))
	}

	f, err := os.Create(filepath.Join(dir, "cities.bin"))
	require.NoError(t, err)
	require.NoError(t, geocode.WriteCitiesBin(f, []geocode.City{
		{Name: "TestCity", Lat: 5, Lon: 5, CountryCode: "XX", Admin1Code: "AA", Admin2Code: "001"},
		{Name: "Other", Lat: 50, Lon: 50, CountryCode: "YY"},
	}))
	require.NoError(t, f.Close())

	ds, err := Load(dir)
	require.NoError(t, err)
	assert.Len(t, ds.Cities, 2)
	assert.Equal(t, 1, len(ds.Countries.Features))
	assert.NotNil(t, ds.KDTree)
	assert.Equal(t, "Test Admin1", ds.Admin1["XX.AA"])
	assert.Equal(t, "Test Admin2", ds.Admin2["XX.AA.001"])
	assert.False(t, ds.UpdatedAt.IsZero())
	assert.Equal(t, 2, ds.CityCount())
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(t.TempDir())
	assert.Error(t, err)
}
