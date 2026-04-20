package updatedata

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// langPriority returns a positive integer indicating how preferable the
// `actual` GeoNames language tag is as a substitute for the requested `desired`
// tag. Lower positive value = better. Zero = reject.
//
// For desired = "zh" we widen to accept related Simplified-Chinese tags
// (GeoNames inconsistently uses zh / zh-Hans / zh-CN / zh-SG / zh-MY).
// Traditional variants (zh-Hant, zh-TW, zh-HK, zh-MO) are intentionally
// skipped so our output stays Simplified.
func langPriority(desired, actual string) int {
	if desired == actual {
		return 1
	}
	if desired == "zh" {
		switch actual {
		case "zh-Hans":
			return 2
		case "zh-CN":
			return 3
		case "zh-SG":
			return 4
		case "zh-MY":
			return 5
		}
	}
	return 0
}

// FilterAlternateNamesByLang streams GeoNames alternateNamesV2.txt and returns
// a map from geonameID → name, limited to entries whose language matches `lang`
// (or a close variant for "zh" — see langPriority) and whose geonameID is
// present in `wanted`. Historic and colloquial entries are skipped.
//
// Selection precedence for a single geonameID:
//  1. Lower langPriority (closer to the requested language family)
//  2. isPreferredName = 1
//  3. isShortName = 1
//  4. First seen
func FilterAlternateNamesByLang(r io.Reader, lang string, wanted map[uint32]bool) (map[uint32]string, error) {
	type candidate struct {
		name      string
		langPrio  int
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
		prio := langPriority(lang, fields[2])
		if prio == 0 {
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

		newCand := candidate{name: name, langPrio: prio, preferred: preferred, short: short}
		cur, seen := best[id]
		if !seen || betterCandidate(newCand, cur) {
			best[id] = newCand
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

// betterCandidate reports whether `n` strictly beats `c` under the
// language-first, then preferred, then short, then first-seen ordering.
func betterCandidate(n, c struct {
	name      string
	langPrio  int
	preferred bool
	short     bool
}) bool {
	if n.langPrio != c.langPrio {
		return n.langPrio < c.langPrio
	}
	if n.preferred != c.preferred {
		return n.preferred && !c.preferred
	}
	if n.short != c.short {
		return n.short && !c.short
	}
	return false
}
