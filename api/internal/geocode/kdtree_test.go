package geocode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKDTreeNearest(t *testing.T) {
	cities := []City{
		{Name: "Berlin", Lat: 52.52, Lon: 13.40, CountryCode: "DE"},
		{Name: "Paris", Lat: 48.85, Lon: 2.35, CountryCode: "FR"},
		{Name: "Tokyo", Lat: 35.68, Lon: 139.65, CountryCode: "JP"},
		{Name: "NewYork", Lat: 40.71, Lon: -74.00, CountryCode: "US"},
	}
	tree := BuildKDTree(cities)
	require.NotNil(t, tree)

	c, ok := tree.Nearest(52.5, 13.4)
	require.True(t, ok)
	assert.Equal(t, "Berlin", c.Name)

	c, ok = tree.Nearest(48.0, 2.0)
	require.True(t, ok)
	assert.Equal(t, "Paris", c.Name)

	c, ok = tree.Nearest(36.0, 140.0)
	require.True(t, ok)
	assert.Equal(t, "Tokyo", c.Name)
}

func TestKDTreeEmpty(t *testing.T) {
	tree := BuildKDTree(nil)
	_, ok := tree.Nearest(0, 0)
	assert.False(t, ok)
}

func TestKDTreeSingleton(t *testing.T) {
	tree := BuildKDTree([]City{{Name: "OnlyOne", Lat: 10, Lon: 10}})
	c, ok := tree.Nearest(999, 999)
	require.True(t, ok)
	assert.Equal(t, "OnlyOne", c.Name)
}

// Brute-force oracle: ensure k-d tree agrees with linear scan.
func TestKDTreeAgreesWithBruteForce(t *testing.T) {
	cities := []City{
		{Name: "A", Lat: 0, Lon: 0},
		{Name: "B", Lat: 1, Lon: 1},
		{Name: "C", Lat: -1, Lon: -1},
		{Name: "D", Lat: 0.5, Lon: 0.5},
		{Name: "E", Lat: 10, Lon: 10},
	}
	tree := BuildKDTree(cities)

	queries := [][2]float64{{0.4, 0.4}, {-0.9, -0.9}, {9, 11}, {100, 100}}
	for _, q := range queries {
		got, ok := tree.Nearest(q[0], q[1])
		require.True(t, ok)

		var best *City
		var bestD float64
		for i := range cities {
			dlat := float64(cities[i].Lat) - q[0]
			dlon := float64(cities[i].Lon) - q[1]
			d := dlat*dlat + dlon*dlon
			if best == nil || d < bestD {
				bestD = d
				best = &cities[i]
			}
		}
		assert.Equal(t, best.Name, got.Name, "query %v", q)
	}
}
