package updatedata

import (
	"bufio"
	"io"
	"strings"
)

// ParseAdminFile parses GeoNames admin1CodesASCII.txt or admin2Codes.txt.
// Format: code<TAB>name<TAB>asciiname<TAB>geonameid (name = column 1, preferred display).
func ParseAdminFile(r io.Reader) (map[string]string, error) {
	m := map[string]string{}
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
		m[fields[0]] = fields[1]
	}
	return m, sc.Err()
}
