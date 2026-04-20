package updatedata

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/wh0am1i/maidenmap/api/internal/geocode"
)

// ParseCitiesGeoNames parses the GeoNames cities15000.txt tab-separated format.
func ParseCitiesGeoNames(r io.Reader) ([]geocode.City, error) {
	var cities []geocode.City
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
		lat, err := strconv.ParseFloat(fields[4], 32)
		if err != nil {
			continue
		}
		lon, err := strconv.ParseFloat(fields[5], 32)
		if err != nil {
			continue
		}
		cities = append(cities, geocode.City{
			Name:        fields[1],
			Lat:         float32(lat),
			Lon:         float32(lon),
			CountryCode: fields[8],
			Admin1Code:  fields[10],
			Admin2Code:  fields[11],
		})
	}
	return cities, sc.Err()
}
