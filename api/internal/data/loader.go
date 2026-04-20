// Package data loads the three offline datasets from disk into in-memory form.
package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/paulmach/orb/geojson"
	"github.com/wh0am1i/maidenmap/api/internal/geocode"
)

// Dataset holds all in-memory data needed to serve geocoding requests.
type Dataset struct {
	Countries       *geojson.FeatureCollection
	CountriesByCode map[string]geocode.Country // ISO alpha-2 → country struct, for code-based fallback
	Cities          []geocode.City
	KDTree          *geocode.KDTree
	Admin1          map[string]geocode.AdminEntry // "US.CA" -> {En, Zh}
	Admin2          map[string]geocode.AdminEntry // "US.CA.037" -> {En, Zh}
	UpdatedAt       time.Time                     // latest mtime of the three files
}

// CityCount returns the number of cities loaded.
func (d *Dataset) CityCount() int { return len(d.Cities) }

type adminCodes struct {
	Admin1 map[string]geocode.AdminEntry `json:"admin1"`
	Admin2 map[string]geocode.AdminEntry `json:"admin2"`
}

// Load reads countries.geojson, cities.bin, admin_codes.json from dir.
func Load(dir string) (*Dataset, error) {
	countries, countriesMtime, err := loadCountries(filepath.Join(dir, "countries.geojson"))
	if err != nil {
		return nil, fmt.Errorf("load countries: %w", err)
	}
	cities, citiesMtime, err := loadCities(filepath.Join(dir, "cities.bin"))
	if err != nil {
		return nil, fmt.Errorf("load cities: %w", err)
	}
	admin, adminMtime, err := loadAdmin(filepath.Join(dir, "admin_codes.json"))
	if err != nil {
		return nil, fmt.Errorf("load admin: %w", err)
	}

	updated := countriesMtime
	if citiesMtime.After(updated) {
		updated = citiesMtime
	}
	if adminMtime.After(updated) {
		updated = adminMtime
	}

	byCode := make(map[string]geocode.Country, len(countries.Features))
	for _, f := range countries.Features {
		code, _ := f.Properties["iso_a2"].(string)
		if code == "" {
			continue
		}
		nameEn, _ := f.Properties["name_en"].(string)
		nameZh, _ := f.Properties["name_zh"].(string)
		byCode[code] = geocode.Country{Code: code, Name: nameEn, NameZh: nameZh}
	}

	return &Dataset{
		Countries:       countries,
		CountriesByCode: byCode,
		Cities:          cities,
		KDTree:          geocode.BuildKDTree(cities),
		Admin1:          admin.Admin1,
		Admin2:          admin.Admin2,
		UpdatedAt:       updated,
	}, nil
}

func loadCountries(path string) (*geojson.FeatureCollection, time.Time, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, time.Time{}, err
	}
	fc, err := geojson.UnmarshalFeatureCollection(b)
	if err != nil {
		return nil, time.Time{}, err
	}
	info, _ := os.Stat(path)
	return fc, info.ModTime(), nil
}

func loadCities(path string) ([]geocode.City, time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer f.Close()
	cities, err := geocode.ReadCitiesBin(f)
	if err != nil {
		return nil, time.Time{}, err
	}
	info, _ := f.Stat()
	return cities, info.ModTime(), nil
}

func loadAdmin(path string) (adminCodes, time.Time, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return adminCodes{}, time.Time{}, err
	}
	var a adminCodes
	if err := json.Unmarshal(b, &a); err != nil {
		return adminCodes{}, time.Time{}, err
	}
	info, _ := os.Stat(path)
	return a, info.ModTime(), nil
}
