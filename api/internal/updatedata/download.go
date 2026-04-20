// Package updatedata contains helpers used by cmd/update-data to fetch and process raw source data.
package updatedata

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// httpClient is tuned for the multi-gigabyte source files we pull (OSM China
// extract ~1.5 GB, alternateNamesV2 ~500 MB). We bound the phases that should
// be fast (dial, TLS, response-header) aggressively, but leave the body copy
// with a generous overall ceiling so slow connections don't blow up mid-file.
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

// DownloadTo fetches url and writes the body to dst (overwriting). If url
// starts with "file://" or looks like an absolute/relative filesystem path
// (no scheme), it's treated as a local copy — useful when operators have
// pre-downloaded the source via a faster channel.
func DownloadTo(url, dst string) error {
	if path, ok := localPath(url); ok {
		return copyFile(path, dst)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "maidenmap-update-data/1.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	return nil
}

// localPath returns (path, true) if url points at a local file.
func localPath(url string) (string, bool) {
	switch {
	case strings.HasPrefix(url, "file://"):
		return strings.TrimPrefix(url, "file://"), true
	case strings.HasPrefix(url, "/"), strings.HasPrefix(url, "./"), strings.HasPrefix(url, "../"):
		return url, true
	}
	return "", false
}

func copyFile(src, dst string) error {
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
