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
	// Tokyo: 东京 is short-name (col 5 = "1"), no preferred present → wins over 东京都.
	// 东京旧名 has isHistoric = 1 (col 7), 东京方言 has isColloquial = 1 (col 6);
	// both should be skipped entirely.
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

func TestFilterAlternateNamesAcceptsZhCNWhenNoBareZh(t *testing.T) {
	f, err := os.Open("testdata/alternate_names_sample.txt")
	require.NoError(t, err)
	defer f.Close()

	// 4000000 only has a zh-CN entry in the fixture; pre-widening it would
	// have been skipped. Now it must come through.
	out, err := FilterAlternateNamesByLang(f, "zh", map[uint32]bool{4000000: true})
	require.NoError(t, err)
	assert.Equal(t, "某城", out[4000000])
}

func TestFilterAlternateNamesSkipsTraditionalInFavorOfSimplified(t *testing.T) {
	f, err := os.Open("testdata/alternate_names_sample.txt")
	require.NoError(t, err)
	defer f.Close()

	// 4000001 has both a zh-TW preferred entry and a plain zh entry.
	// zh-TW must be rejected outright; the zh entry wins.
	out, err := FilterAlternateNamesByLang(f, "zh", map[uint32]bool{4000001: true})
	require.NoError(t, err)
	assert.Equal(t, "某镇", out[4000001])
}

func TestFilterAlternateNamesLangPriorityThenPreferred(t *testing.T) {
	f, err := os.Open("testdata/alternate_names_sample.txt")
	require.NoError(t, err)
	defer f.Close()

	// 4000002 has:
	//   - zh-Hans preferred=1  → priority 2, preferred
	//   - zh-CN   preferred=0  → priority 3
	// zh-Hans outranks on language priority alone; preferred flag confirms.
	out, err := FilterAlternateNamesByLang(f, "zh", map[uint32]bool{4000002: true})
	require.NoError(t, err)
	assert.Equal(t, "某区", out[4000002])
}
