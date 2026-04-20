package geocode

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCitiesBinRoundTrip(t *testing.T) {
	in := []City{
		{Name: "Cottbus", NameZh: "科特布斯", Lat: 51.76, Lon: 14.33, CountryCode: "DE", Admin1Code: "BR", Admin2Code: "12071"},
		{Name: "Beijing", NameZh: "北京", Lat: 39.9042, Lon: 116.4074, CountryCode: "CN", Admin1Code: "22"},
		{Name: "Tokyo", NameZh: "", Lat: 35.6762, Lon: 139.6503, CountryCode: "JP", Admin1Code: "40"},
	}

	var buf bytes.Buffer
	require.NoError(t, WriteCitiesBin(&buf, in))

	out, err := ReadCitiesBin(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	require.Len(t, out, 3)
	for i, c := range in {
		assert.Equal(t, c.Name, out[i].Name)
		assert.Equal(t, c.NameZh, out[i].NameZh)
		assert.InDelta(t, c.Lat, out[i].Lat, 1e-4)
		assert.InDelta(t, c.Lon, out[i].Lon, 1e-4)
		assert.Equal(t, c.CountryCode, out[i].CountryCode)
		assert.Equal(t, c.Admin1Code, out[i].Admin1Code)
		assert.Equal(t, c.Admin2Code, out[i].Admin2Code)
	}
}

func TestCitiesBinBadMagic(t *testing.T) {
	buf := bytes.NewReader([]byte("XXXX\x02\x00\x00\x00\x00\x00\x00\x00"))
	_, err := ReadCitiesBin(buf)
	assert.Error(t, err)
}

func TestCitiesBinRejectsV1(t *testing.T) {
	// magic "MMCB" + version 1 + count 0
	data := []byte{'M', 'M', 'C', 'B', 1, 0, 0, 0, 0, 0, 0, 0}
	_, err := ReadCitiesBin(bytes.NewReader(data))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rerun update-data")
}

func TestCitiesBinBadVersion(t *testing.T) {
	data := []byte{'M', 'M', 'C', 'B', 99, 0, 0, 0, 0, 0, 0, 0}
	_, err := ReadCitiesBin(bytes.NewReader(data))
	assert.Error(t, err)
}

func TestCitiesBinRejectsAbsurdCount(t *testing.T) {
	// Magic + version 2 + count 0xFFFFFFFF. Without the cap this would
	// attempt to pre-allocate ~200 GB.
	data := []byte{'M', 'M', 'C', 'B', 2, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF}
	_, err := ReadCitiesBin(bytes.NewReader(data))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds sanity cap")
}

func TestCitiesBinEmpty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteCitiesBin(&buf, nil))
	out, err := ReadCitiesBin(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Empty(t, out)
}
