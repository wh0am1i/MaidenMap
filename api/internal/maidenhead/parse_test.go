package maidenhead

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse4Char(t *testing.T) {
	// JO65 center: lon = -180 + 9*20 + 6*2 + 1 = 13, lat = -90 + 14*10 + 5*1 + 0.5 = 55.5
	loc, err := Parse("JO65")
	require.NoError(t, err)
	assert.Equal(t, "JO65", loc.Grid)
	assert.InDelta(t, 13.0, loc.Lon, 1e-6)
	assert.InDelta(t, 55.5, loc.Lat, 1e-6)
}

func TestParse6Char(t *testing.T) {
	// JO65ab: base JO65 SW corner (12, 55), + 0*5' lon, 1*2.5' lat, +half subsquare
	// lon = 12 + 0/12 + 1/24 ≈ 12.04167
	// lat = 55 + 1/24 + 1/48 ≈ 55.0625
	loc, err := Parse("JO65ab")
	require.NoError(t, err)
	assert.InDelta(t, 12.0+0.0/12.0+1.0/24.0, loc.Lon, 1e-6)
	assert.InDelta(t, 55.0+1.0/24.0+1.0/48.0, loc.Lat, 1e-6)
}

func TestParse8Char(t *testing.T) {
	loc, err := Parse("JO65ab12")
	require.NoError(t, err)
	assert.Equal(t, "JO65ab12", loc.Grid)
	// Just assert it's inside the 6-char subsquare
	assert.InDelta(t, 12.01, loc.Lon, 0.1)
	assert.InDelta(t, 55.05, loc.Lat, 0.1)
}

func TestParseLowerCaseFirstPair(t *testing.T) {
	// 首两位也接受小写，但 subsquare 固定 a-x 小写
	loc, err := Parse("jo65")
	require.NoError(t, err)
	assert.InDelta(t, 13.0, loc.Lon, 1e-6)
	assert.InDelta(t, 55.5, loc.Lat, 1e-6)
}

func TestParseCorners(t *testing.T) {
	// AA00 center = (-180 + 0*20 + 0*2 + 1, -90 + 0*10 + 0*1 + 0.5) = (-179, -89.5)
	loc, err := Parse("AA00")
	require.NoError(t, err)
	assert.InDelta(t, -179.0, loc.Lon, 1e-6)
	assert.InDelta(t, -89.5, loc.Lat, 1e-6)

	// RR99 center = (-180 + 17*20 + 9*2 + 1, -90 + 17*10 + 9*1 + 0.5) = (179, 89.5)
	loc, err = Parse("RR99")
	require.NoError(t, err)
	assert.InDelta(t, 179.0, loc.Lon, 1e-6)
	assert.InDelta(t, 89.5, loc.Lat, 1e-6)
}

func TestParseInvalid(t *testing.T) {
	cases := []string{
		"",          // empty
		"J",         // too short
		"JO6",       // odd length
		"JO6500ab",  // 8 char must end with digits
		"ZZ00",      // field out of range (S-Z invalid)
		"JO65YY",    // subsquare must be a-x
		"JO65ab1",   // odd length
		"1234",      // digits in field position
		"JOJO",      // letters in square position
	}
	for _, c := range cases {
		_, err := Parse(c)
		assert.Error(t, err, "expected error for %q", c)
	}
}
