package updatedata

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// FilterAlternateNamesByLang streams GeoNames alternateNamesV2.txt and returns
// a map from geonameID → name, limited to entries whose language matches `lang`
// and whose geonameID is present in `wanted`. Historic and colloquial entries
// are skipped. Selection priority: preferred > short > first seen.
func FilterAlternateNamesByLang(r io.Reader, lang string, wanted map[uint32]bool) (map[uint32]string, error) {
	type candidate struct {
		name      string
		preferred bool
		short     bool
	}
	best := map[uint32]candidate{}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			continue
		}
		if fields[2] != lang {
			continue
		}
		id64, err := strconv.ParseUint(fields[1], 10, 32)
		if err != nil {
			continue
		}
		id := uint32(id64)
		if !wanted[id] {
			continue
		}
		name := fields[3]
		if name == "" {
			continue
		}

		preferred := len(fields) > 4 && fields[4] == "1"
		short := len(fields) > 5 && fields[5] == "1"
		colloquial := len(fields) > 6 && fields[6] == "1"
		historic := len(fields) > 7 && fields[7] == "1"
		if colloquial || historic {
			continue
		}

		cur, seen := best[id]
		if !seen {
			best[id] = candidate{name: name, preferred: preferred, short: short}
			continue
		}
		// Replace only if new is strictly better:
		// preferred beats non-preferred; among non-preferred, short beats non-short.
		if preferred && !cur.preferred {
			best[id] = candidate{name: name, preferred: true, short: short}
			continue
		}
		if !cur.preferred && short && !cur.short {
			best[id] = candidate{name: name, preferred: preferred, short: true}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	out := make(map[uint32]string, len(best))
	for id, c := range best {
		out[id] = c.name
	}
	return out, nil
}
