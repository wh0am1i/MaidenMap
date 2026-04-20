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

func TestDownloadToRejectsBarePath(t *testing.T) {
	// Bare absolute/relative paths used to be accepted as "local copy". That
	// was a footgun (--countries-url=/etc/passwd would silently be copied).
	// Now we require an explicit file:// scheme.
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	require.NoError(t, os.WriteFile(src, []byte("local-data"), 0o644))

	dst := filepath.Join(dir, "dst.bin")
	err := DownloadTo(src, dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scheme")
}

func TestDownloadToRejectsUnsupportedScheme(t *testing.T) {
	err := DownloadTo("ftp://example.com/foo", "/tmp/anything")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scheme")
}

func TestDownloadToBodyCapExceeded(t *testing.T) {
	orig := maxDownloadBytes
	maxDownloadBytes = 128
	t.Cleanup(func() { maxDownloadBytes = orig })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(make([]byte, 1024)) // well past the 128-byte cap
	}))
	defer srv.Close()

	dst := filepath.Join(t.TempDir(), "out.bin")
	err := DownloadTo(srv.URL, dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds cap")
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
