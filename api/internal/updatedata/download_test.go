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

func TestDownloadToLocalPath(t *testing.T) {
	// Operators can pre-download via a fast channel and pass an absolute
	// path instead of a URL. DownloadTo should copy the file.
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	require.NoError(t, os.WriteFile(src, []byte("local-data"), 0o644))

	dst := filepath.Join(dir, "dst.bin")
	require.NoError(t, DownloadTo(src, dst))

	b, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "local-data", string(b))
}

func TestDownloadToFileURL(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	require.NoError(t, os.WriteFile(src, []byte("via-file-url"), 0o644))

	dst := filepath.Join(dir, "dst.bin")
	require.NoError(t, DownloadTo("file://"+src, dst))

	b, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "via-file-url", string(b))
}
