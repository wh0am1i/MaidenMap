package updatedata

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterAlternateNamesPreferred(t *testing.T) {
	f, err := os.Open("testdata/alternate_names_sample.txt")
	require.NoError(t, err)
	defer f.Close()

	// Include Berlin (2950159), Tokyo (1850147), Berlin admin1 (2950157).
	// Exclude 9999999.
	wanted := map[uint32]bool{
		2950159: true,
		1850147: true,
		2950157: true,
	}
	out, err := FilterAlternateNamesByLang(f, "zh", wanted)
	require.NoError(t, err)

	// Berlin: 柏林 is preferred (col 4 = "1") → wins over 柏林市
	assert.Equal(t, "柏林", out[2950159])
	// Tokyo: 东京 is short-name (col 5 = "1"), no preferred present → wins over 东京都
	// 东京旧名 has isHistoric = 1, should be skipped entirely
	assert.Equal(t, "东京", out[1850147])
	// Berlin admin1: only one match
	assert.Equal(t, "柏林州", out[2950157])
	// Not wanted should be absent
	_, hasUnwanted := out[9999999]
	assert.False(t, hasUnwanted)
}

func TestFilterAlternateNamesWrongLang(t *testing.T) {
	f, err := os.Open("testdata/alternate_names_sample.txt")
	require.NoError(t, err)
	defer f.Close()

	// Fetch "en" names instead; should get Berlin = "Berlin", nothing for Tokyo.
	wanted := map[uint32]bool{2950159: true, 1850147: true}
	out, err := FilterAlternateNamesByLang(f, "en", wanted)
	require.NoError(t, err)

	assert.Equal(t, "Berlin", out[2950159])
	_, hasTokyo := out[1850147]
	assert.False(t, hasTokyo)
}
