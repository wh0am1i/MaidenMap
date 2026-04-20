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
	assert.Equal(t, "Berlin", m["DE.16"])
	assert.Equal(t, "Tokyo", m["JP.40"])
}
