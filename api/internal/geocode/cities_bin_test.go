package geocode

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCitiesBinRoundTrip(t *testing.T) {
	in := []City{
		{Name: "Cottbus", Lat: 51.76, Lon: 14.33, CountryCode: "DE", Admin1Code: "BR", Admin2Code: "12071"},
		{Name: "北京", Lat: 39.9042, Lon: 116.4074, CountryCode: "CN", Admin1Code: "22", Admin2Code: ""},
		{Name: "Tokyo", Lat: 35.6762, Lon: 139.6503, CountryCode: "JP", Admin1Code: "40", Admin2Code: ""},
	}

	var buf bytes.Buffer
	require.NoError(t, WriteCitiesBin(&buf, in))

	out, err := ReadCitiesBin(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	require.Len(t, out, 3)
	for i, c := range in {
		assert.Equal(t, c.Name, out[i].Name)
		assert.InDelta(t, c.Lat, out[i].Lat, 1e-4)
		assert.InDelta(t, c.Lon, out[i].Lon, 1e-4)
		assert.Equal(t, c.CountryCode, out[i].CountryCode)
		assert.Equal(t, c.Admin1Code, out[i].Admin1Code)
		assert.Equal(t, c.Admin2Code, out[i].Admin2Code)
	}
}

func TestCitiesBinBadMagic(t *testing.T) {
	buf := bytes.NewReader([]byte("XXXX\x01\x00\x00\x00\x00\x00\x00\x00"))
	_, err := ReadCitiesBin(buf)
	assert.Error(t, err)
}

func TestCitiesBinBadVersion(t *testing.T) {
	// magic "MMCB" + version 99 + count 0
	data := []byte{'M', 'M', 'C', 'B', 99, 0, 0, 0, 0, 0, 0, 0}
	_, err := ReadCitiesBin(bytes.NewReader(data))
	assert.Error(t, err)
}

func TestCitiesBinEmpty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteCitiesBin(&buf, nil))
	out, err := ReadCitiesBin(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Empty(t, out)
}
