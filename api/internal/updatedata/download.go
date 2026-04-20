// Package updatedata contains helpers used by cmd/update-data to fetch and process raw source data.
package updatedata

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// httpClient is tuned for the multi-hundred-MB source files we pull
// (alternateNamesV2 ~500 MB, cities15000 ~30 MB). Fast phases (dial, TLS,
// response-header) are bound aggressively; the total ceiling is generous so
// slow connections don't fail mid-file.
var httpClient = &http.Client{
	Timeout: 30 * time.Minute,
	Transport: &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	},
}

// maxDownloadBytes caps any single source download. The largest real source
// today is alternateNamesV2 at ~500 MB; 2 GB leaves plenty of headroom while
// bounding OOM / disk-fill risk from a compromised or misbehaving upstream
// that streams arbitrary data into os.Create. Overridable in tests.
var maxDownloadBytes int64 = 2 << 30 // 2 GiB

// DownloadTo fetches url and writes the body to dst (overwriting).
//
// Supported schemes:
//   - https:// — preferred for any remote source.
//   - http://  — permitted for localhost/test servers; callers should use
//     https for production sources.
//   - file://  — local file copy, useful when operators have pre-downloaded
//     a source via a faster channel. The path after file:// is resolved as-is.
//
// Bare paths ("/etc/passwd", "../secret") are rejected — operators who want
// a local copy must spell it file:///path.
func DownloadTo(rawURL, dst string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse url %q: %w", rawURL, err)
	}
	switch u.Scheme {
	case "file":
		return copyFile(u.Path, dst)
	case "http", "https":
		return downloadHTTP(rawURL, dst)
	default:
		return fmt.Errorf("unsupported url scheme %q (want https / http / file)", u.Scheme)
	}
}

func downloadHTTP(rawURL, dst string) error {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "maidenmap-update-data/1.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("GET %s: status %d", rawURL, resp.StatusCode)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	// LimitReader + one extra byte lets us distinguish "exactly at cap" from
	// "upstream kept streaming past the cap" and fail the latter.
	n, err := io.Copy(f, io.LimitReader(resp.Body, maxDownloadBytes+1))
	if err != nil {
		return err
	}
	if n > maxDownloadBytes {
		return fmt.Errorf("GET %s: body exceeds cap (%d bytes)", rawURL, maxDownloadBytes)
	}
	return nil
}

func copyFile(src, dst string) error {
	if src == "" {
		return fmt.Errorf("file:// url missing path")
	}
	// Guard against obvious footguns; operators can still target arbitrary
	// readable paths but at least can't pass a bare "/etc/passwd" that looked
	// like a remote URL by accident.
	if strings.ContainsRune(src, '\x00') {
		return fmt.Errorf("invalid file path")
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
