// Command maidenmap-update-data fetches raw GeoNames + Natural Earth + DataV
// data and writes the consumed data files atomically.
package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/longbridgeapp/opencc"
	"github.com/wh0am1i/maidenmap/api/internal/geocode"
	"github.com/wh0am1i/maidenmap/api/internal/updatedata"
)

const (
	defaultCitiesURL         = "https://download.geonames.org/export/dump/cities15000.zip"
	defaultAdmin1URL         = "https://download.geonames.org/export/dump/admin1CodesASCII.txt"
	defaultAdmin2URL         = "https://download.geonames.org/export/dump/admin2Codes.txt"
	defaultCountriesURL      = "https://raw.githubusercontent.com/nvkelso/natural-earth-vector/master/geojson/ne_10m_admin_0_countries.geojson"
	defaultAlternateNamesURL = "https://download.geonames.org/export/dump/alternateNamesV2.zip"

	minCities    = 10000
	minCountries = 150

	// DataV drill concurrency — one HTTP call per province, then one per
	// mainland city. 8 keeps the Alibaba endpoint happy.
	dataVConcurrency = 8
)

func main() {
	dataDir := flag.String("data-dir", envDefault("DATA_DIR", "./data"), "output directory")
	citiesURL := flag.String("cities-url", envDefault("CITIES_URL", defaultCitiesURL), "")
	admin1URL := flag.String("admin1-url", envDefault("ADMIN1_URL", defaultAdmin1URL), "")
	admin2URL := flag.String("admin2-url", envDefault("ADMIN2_URL", defaultAdmin2URL), "")
	countriesURL := flag.String("countries-url", envDefault("COUNTRIES_URL", defaultCountriesURL), "")
	altNamesURL := flag.String("alt-names-url", envDefault("ALT_NAMES_URL", defaultAlternateNamesURL), "")
	dataVURL := flag.String("datav-url", envDefault("DATAV_URL", updatedata.DefaultDataVBaseURL), "DataV base URL; empty disables DataV fetch")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	tmp, err := os.MkdirTemp("", "maidenmap-update-*")
	if err != nil {
		fatal("mkdir temp", err)
	}
	defer os.RemoveAll(tmp)

	slog.Info("download", "what", "cities")
	citiesZip := filepath.Join(tmp, "cities.zip")
	must(updatedata.DownloadTo(*citiesURL, citiesZip), "cities")

	slog.Info("download", "what", "admin1")
	admin1Path := filepath.Join(tmp, "admin1.txt")
	must(updatedata.DownloadTo(*admin1URL, admin1Path), "admin1")

	slog.Info("download", "what", "admin2")
	admin2Path := filepath.Join(tmp, "admin2.txt")
	must(updatedata.DownloadTo(*admin2URL, admin2Path), "admin2")

	slog.Info("download", "what", "countries")
	countriesPath := filepath.Join(tmp, "countries.geojson")
	must(updatedata.DownloadTo(*countriesURL, countriesPath), "countries")

	slog.Info("download", "what", "alt-names")
	altZip := filepath.Join(tmp, "alternateNamesV2.zip")
	must(updatedata.DownloadTo(*altNamesURL, altZip), "alt-names")

	citiesTxt := filepath.Join(tmp, "cities15000.txt")
	must(unzipSingle(citiesZip, "cities15000.txt", citiesTxt), "unzip cities")

	cf, err := os.Open(citiesTxt)
	if err != nil {
		fatal("open cities.txt", err)
	}
	cityEntries, err := updatedata.ParseCitiesGeoNames(cf)
	cf.Close()
	if err != nil {
		fatal("parse cities", err)
	}
	if len(cityEntries) < minCities {
		fatal("too few cities", fmt.Errorf("got %d, need >= %d", len(cityEntries), minCities))
	}
	slog.Info("parsed cities", "count", len(cityEntries))

	a1Raw, err := parseAdminFrom(admin1Path)
	if err != nil {
		fatal("parse admin1", err)
	}
	a2Raw, err := parseAdminFrom(admin2Path)
	if err != nil {
		fatal("parse admin2", err)
	}
	slog.Info("parsed admin", "admin1", len(a1Raw), "admin2", len(a2Raw))

	countriesRaw, err := os.ReadFile(countriesPath)
	if err != nil {
		fatal("read countries", err)
	}
	fc, err := updatedata.ParseNaturalEarthCountries(countriesRaw)
	if err != nil {
		fatal("parse countries", err)
	}
	if len(fc.Features) < minCountries {
		fatal("too few countries", fmt.Errorf("got %d, need >= %d", len(fc.Features), minCountries))
	}
	slog.Info("parsed countries", "count", len(fc.Features))

	wanted := make(map[uint32]bool, len(cityEntries)+len(a1Raw)+len(a2Raw))
	for _, ce := range cityEntries {
		wanted[ce.GeonameID] = true
	}
	for _, e := range a1Raw {
		if e.GeonameID != 0 {
			wanted[e.GeonameID] = true
		}
	}
	for _, e := range a2Raw {
		if e.GeonameID != 0 {
			wanted[e.GeonameID] = true
		}
	}

	altTxt := filepath.Join(tmp, "alternateNamesV2.txt")
	must(unzipSingle(altZip, "alternateNamesV2.txt", altTxt), "unzip alt-names")

	af, err := os.Open(altTxt)
	if err != nil {
		fatal("open alt-names.txt", err)
	}
	zhByID, err := updatedata.FilterAlternateNamesByLang(af, "zh", wanted)
	af.Close()
	if err != nil {
		fatal("filter alt-names", err)
	}
	slog.Info("filtered zh names", "count", len(zhByID))

	// GeoNames tags some Traditional Chinese names as plain "zh", so the
	// filter alone can't guarantee Simplified output. Run everything through
	// OpenCC t2s; converting already-Simplified text is idempotent.
	t2s, err := opencc.New("t2s")
	if err != nil {
		fatal("init opencc t2s", err)
	}
	converted := 0
	for id, name := range zhByID {
		out, err := t2s.Convert(name)
		if err != nil || out == "" {
			continue
		}
		if out != name {
			converted++
		}
		zhByID[id] = out
	}
	slog.Info("converted zh to simplified", "changed", converted)

	// Also run country name_zh through t2s (Natural Earth is mostly Simplified
	// but not guaranteed; HK/MO/TW overrides are already Simplified — idempotent).
	for _, feat := range fc.Features {
		nz, ok := feat.Properties["name_zh"].(string)
		if !ok || nz == "" {
			continue
		}
		if out, err := t2s.Convert(nz); err == nil && out != "" {
			feat.Properties["name_zh"] = out
		}
	}

	cities := make([]geocode.City, 0, len(cityEntries))
	for _, ce := range cityEntries {
		c := ce.City
		if zh, ok := zhByID[ce.GeonameID]; ok {
			c.NameZh = zh
		}
		cities = append(cities, c)
	}

	a1Final := make(map[string]geocode.AdminEntry, len(a1Raw))
	for code, e := range a1Raw {
		a1Final[code] = geocode.AdminEntry{En: e.Name, Zh: zhByID[e.GeonameID]}
	}
	a2Final := make(map[string]geocode.AdminEntry, len(a2Raw))
	for code, e := range a2Raw {
		a2Final[code] = geocode.AdminEntry{En: e.Name, Zh: zhByID[e.GeonameID]}
	}

	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		fatal("mkdir data dir", err)
	}

	fcJSON, err := fc.MarshalJSON()
	if err != nil {
		fatal("marshal countries", err)
	}
	must(atomicWrite(filepath.Join(*dataDir, "countries.geojson"), fcJSON), "write countries")

	adminJSON, err := json.Marshal(map[string]any{"admin1": a1Final, "admin2": a2Final})
	if err != nil {
		fatal("marshal admin", err)
	}
	must(atomicWrite(filepath.Join(*dataDir, "admin_codes.json"), adminJSON), "write admin")

	if err := atomicWriteFunc(filepath.Join(*dataDir, "cities.bin"), func(w io.Writer) error {
		return geocode.WriteCitiesBin(w, cities)
	}); err != nil {
		fatal("write cities.bin", err)
	}

	// DataV fetch runs last because it's the slowest step (~400 HTTP calls).
	// Emit to datav.geojson as a slim FeatureCollection — the API loads it
	// optionally and falls back to GeoNames when absent.
	if *dataVURL != "" {
		slog.Info("download", "what", "datav")
		nodes, err := updatedata.FetchDataVChina(*dataVURL, dataVConcurrency)
		if err != nil {
			slog.Warn("datav fetch failed; skipping", "err", err)
		} else {
			if err := atomicWriteFunc(filepath.Join(*dataDir, "datav.geojson"), func(w io.Writer) error {
				return updatedata.EncodeDataVNodes(w, nodes)
			}); err != nil {
				fatal("write datav.geojson", err)
			}
			slog.Info("datav written", "nodes", len(nodes))
		}
	}

	slog.Info("update complete", "data_dir", *dataDir, "cities_with_zh", countNonEmptyZh(cities))
}

func countNonEmptyZh(cities []geocode.City) int {
	n := 0
	for _, c := range cities {
		if c.NameZh != "" {
			n++
		}
	}
	return n
}

func parseAdminFrom(path string) (map[string]updatedata.AdminParseEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return updatedata.ParseAdminFile(f)
}

func unzipSingle(zipPath, entryName, dst string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name != entryName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, rc)
		return err
	}
	return fmt.Errorf("entry %q not found in zip", entryName)
}

func atomicWrite(path string, data []byte) error {
	return atomicWriteFunc(path, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
}

func atomicWriteFunc(path string, fn func(io.Writer) error) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := fn(f); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

func must(err error, label string) {
	if err != nil {
		fatal(label, err)
	}
}

func fatal(label string, err error) {
	slog.Error(label, "err", err)
	os.Exit(1)
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
