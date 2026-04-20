package updatedata

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAdminFile(t *testing.T) {
	f, err := os.Open("testdata/admin1.txt")
	require.NoError(t, err)
	defer f.Close()
	m, err := ParseAdminFile(f)
	require.NoError(t, err)

	require.Contains(t, m, "DE.16")
	assert.Equal(t, "Berlin", m["DE.16"].Name)
	assert.Equal(t, uint32(2950157), m["DE.16"].GeonameID)

	require.Contains(t, m, "JP.40")
	assert.Equal(t, "Tokyo", m["JP.40"].Name)
	assert.Equal(t, uint32(1850144), m["JP.40"].GeonameID)
}
