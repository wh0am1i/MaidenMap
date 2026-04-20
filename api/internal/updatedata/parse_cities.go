package updatedata

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/wh0am1i/maidenmap/api/internal/geocode"
)

// CityEntry pairs a parsed City with its upstream GeoNames ID (used for joining
// alternateNames).
type CityEntry struct {
	GeonameID uint32
	City      geocode.City
}

// ParseCitiesGeoNames parses the GeoNames cities15000.txt tab-separated format.
func ParseCitiesGeoNames(r io.Reader) ([]CityEntry, error) {
	var entries []CityEntry
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 12 {
			continue
		}
		id, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			continue
		}
		lat, err := strconv.ParseFloat(fields[4], 32)
		if err != nil {
			continue
		}
		lon, err := strconv.ParseFloat(fields[5], 32)
		if err != nil {
			continue
		}
		entries = append(entries, CityEntry{
			GeonameID: uint32(id),
			City: geocode.City{
				Name:        fields[1],
				Lat:         float32(lat),
				Lon:         float32(lon),
				CountryCode: fields[8],
				Admin1Code:  fields[10],
				Admin2Code:  fields[11],
			},
		})
	}
	return entries, sc.Err()
}
