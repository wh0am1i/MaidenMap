package updatedata

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// AdminParseEntry is the raw parser output for admin1/admin2: the display name
// plus the upstream geonameID used for joining alternateNames.
type AdminParseEntry struct {
	Name      string
	GeonameID uint32
}

// ParseAdminFile parses GeoNames admin1CodesASCII.txt or admin2Codes.txt.
// Format: code<TAB>name<TAB>asciiname<TAB>geonameid.
func ParseAdminFile(r io.Reader) (map[string]AdminParseEntry, error) {
	m := map[string]AdminParseEntry{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}
		entry := AdminParseEntry{Name: fields[1]}
		if len(fields) >= 4 {
			if id, err := strconv.ParseUint(fields[3], 10, 32); err == nil {
				entry.GeonameID = uint32(id)
			}
		}
		m[fields[0]] = entry
	}
	return m, sc.Err()
}
