package updatedata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNearestOSMPlaceWithin(t *testing.T) {
	places := []OSMPlace{
		{Lat: 30.2741, Lon: 120.1551, NameZh: "杭州"}, // Hangzhou
		{Lat: 30.0489, Lon: 119.9570, NameZh: "富阳"}, // Fuyang (Hangzhou district)
		{Lat: 31.2304, Lon: 121.4737, NameZh: "上海"}, // Shanghai
	}

	// Query near Fuyang — closest is Fuyang, within 5km.
	name, ok := NearestOSMPlaceWithin(places, 30.05, 119.96, 5)
	assert.True(t, ok)
	assert.Equal(t, "富阳", name)

	// Query near Shanghai — Shanghai wins.
	name, ok = NearestOSMPlaceWithin(places, 31.23, 121.47, 5)
	assert.True(t, ok)
	assert.Equal(t, "上海", name)

	// Query in the middle of the ocean — outside 5km of any place.
	_, ok = NearestOSMPlaceWithin(places, 0, 0, 5)
	assert.False(t, ok)
}

func TestNearestOSMPlaceEmpty(t *testing.T) {
	_, ok := NearestOSMPlaceWithin(nil, 30, 120, 5)
	assert.False(t, ok)
}

func TestHaversineKm(t *testing.T) {
	// Hangzhou to Shanghai: known roughly 160 km
	d := haversineKm(30.27, 120.15, 31.23, 121.47)
	assert.InDelta(t, 165, d, 10)

	// Same point
	assert.InDelta(t, 0, haversineKm(30, 120, 30, 120), 0.001)
}
