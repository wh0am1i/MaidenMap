// Command maidenmap-update-data fetches raw GeoNames + Natural Earth data and
// writes the three consumed data files atomically.
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

	"github.com/wh0am1i/maidenmap/api/internal/geocode"
	"github.com/wh0am1i/maidenmap/api/internal/updatedata"
)

const (
	defaultCitiesURL    = "https://download.geonames.org/export/dump/cities15000.zip"
	defaultAdmin1URL    = "https://download.geonames.org/export/dump/admin1CodesASCII.txt"
	defaultAdmin2URL    = "https://download.geonames.org/export/dump/admin2Codes.txt"
	defaultCountriesURL = "https://raw.githubusercontent.com/nvkelso/natural-earth-vector/master/geojson/ne_50m_admin_0_countries.geojson"

	minCities    = 10000
	minCountries = 150
)

func main() {
	dataDir := flag.String("data-dir", envDefault("DATA_DIR", "./data"), "output directory")
	citiesURL := flag.String("cities-url", envDefault("CITIES_URL", defaultCitiesURL), "")
	admin1URL := flag.String("admin1-url", envDefault("ADMIN1_URL", defaultAdmin1URL), "")
	admin2URL := flag.String("admin2-url", envDefault("ADMIN2_URL", defaultAdmin2URL), "")
	countriesURL := flag.String("countries-url", envDefault("COUNTRIES_URL", defaultCountriesURL), "")
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

	citiesTxt := filepath.Join(tmp, "cities15000.txt")
	must(unzipSingle(citiesZip, "cities15000.txt", citiesTxt), "unzip cities")

	cf, err := os.Open(citiesTxt)
	if err != nil {
		fatal("open cities.txt", err)
	}
	cities, err := updatedata.ParseCitiesGeoNames(cf)
	cf.Close()
	if err != nil {
		fatal("parse cities", err)
	}
	if len(cities) < minCities {
		fatal("too few cities", fmt.Errorf("got %d, need >= %d", len(cities), minCities))
	}
	slog.Info("parsed cities", "count", len(cities))

	a1, err := parseAdminFrom(admin1Path)
	if err != nil {
		fatal("parse admin1", err)
	}
	a2, err := parseAdminFrom(admin2Path)
	if err != nil {
		fatal("parse admin2", err)
	}
	slog.Info("parsed admin", "admin1", len(a1), "admin2", len(a2))

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

	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		fatal("mkdir data dir", err)
	}

	fcJSON, err := fc.MarshalJSON()
	if err != nil {
		fatal("marshal countries", err)
	}
	must(atomicWrite(filepath.Join(*dataDir, "countries.geojson"), fcJSON), "write countries")

	adminJSON, err := json.Marshal(map[string]any{"admin1": a1, "admin2": a2})
	if err != nil {
		fatal("marshal admin", err)
	}
	must(atomicWrite(filepath.Join(*dataDir, "admin_codes.json"), adminJSON), "write admin")

	if err := atomicWriteFunc(filepath.Join(*dataDir, "cities.bin"), func(w io.Writer) error {
		return geocode.WriteCitiesBin(w, cities)
	}); err != nil {
		fatal("write cities.bin", err)
	}

	slog.Info("update complete", "data_dir", *dataDir)
}

func parseAdminFrom(path string) (map[string]string, error) {
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
