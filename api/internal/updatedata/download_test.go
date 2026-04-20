package updatedata

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadTo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	defer srv.Close()

	dst := filepath.Join(t.TempDir(), "out.txt")
	require.NoError(t, DownloadTo(srv.URL, dst))

	b, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(b))
}

func TestDownloadToNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", 500)
	}))
	defer srv.Close()

	dst := filepath.Join(t.TempDir(), "out.txt")
	err := DownloadTo(srv.URL, dst)
	assert.Error(t, err)
}
