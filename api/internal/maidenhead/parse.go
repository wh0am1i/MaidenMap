// Package maidenhead converts Maidenhead grid locators to geographic coordinates.
package maidenhead

import (
	"fmt"
	"strings"
)

// Location is a parsed Maidenhead locator with its center coordinates.
type Location struct {
	Grid string  // Original grid string (normalized: upper-upper, lower subsquare)
	Lat  float64 // Center latitude (-90..90)
	Lon  float64 // Center longitude (-180..180)
}

// Parse parses a 4/6/8 character Maidenhead locator and returns the center coordinates.
func Parse(raw string) (Location, error) {
	s := strings.TrimSpace(raw)
	n := len(s)
	if n != 4 && n != 6 && n != 8 {
		return Location{}, fmt.Errorf("invalid length %d: must be 4, 6, or 8", n)
	}

	// Normalize: field/subsquare letters to canonical case for validation.
	upper := strings.ToUpper(s[:2])
	mid := s[2:4]
	var norm strings.Builder
	norm.WriteString(upper)
	norm.WriteString(mid)

	// Validate field (chars 0-1): A-R
	for i := 0; i < 2; i++ {
		c := upper[i]
		if c < 'A' || c > 'R' {
			return Location{}, fmt.Errorf("invalid field char %q at position %d", c, i)
		}
	}
	// Validate square (chars 2-3): 0-9
	for i := 0; i < 2; i++ {
		c := mid[i]
		if c < '0' || c > '9' {
			return Location{}, fmt.Errorf("invalid square char %q at position %d", c, i+2)
		}
	}

	lonField := float64(upper[0] - 'A')
	latField := float64(upper[1] - 'A')
	lonSquare := float64(mid[0] - '0')
	latSquare := float64(mid[1] - '0')

	// 4-char center
	lon := -180.0 + lonField*20.0 + lonSquare*2.0 + 1.0
	lat := -90.0 + latField*10.0 + latSquare*1.0 + 0.5

	if n >= 6 {
		sub := strings.ToLower(s[4:6])
		for i := 0; i < 2; i++ {
			c := sub[i]
			if c < 'a' || c > 'x' {
				return Location{}, fmt.Errorf("invalid subsquare char %q at position %d", c, i+4)
			}
		}
		norm.WriteString(sub)
		lonSub := float64(sub[0] - 'a')
		latSub := float64(sub[1] - 'a')
		// Override earlier "+1 / +0.5" (4-char was the center of a 2°×1° box, now subdivide)
		lon = -180.0 + lonField*20.0 + lonSquare*2.0 + lonSub*(5.0/60.0) + (5.0/60.0)/2.0
		lat = -90.0 + latField*10.0 + latSquare*1.0 + latSub*(2.5/60.0) + (2.5/60.0)/2.0
	}

	if n == 8 {
		ext := s[6:8]
		for i := 0; i < 2; i++ {
			c := ext[i]
			if c < '0' || c > '9' {
				return Location{}, fmt.Errorf("invalid extended char %q at position %d", c, i+6)
			}
		}
		norm.WriteString(ext)
		lonExt := float64(ext[0] - '0')
		latExt := float64(ext[1] - '0')
		sub := strings.ToLower(s[4:6])
		lonSub := float64(sub[0] - 'a')
		latSub := float64(sub[1] - 'a')
		lon = -180.0 + lonField*20.0 + lonSquare*2.0 + lonSub*(5.0/60.0) + lonExt*(30.0/3600.0) + (30.0/3600.0)/2.0
		lat = -90.0 + latField*10.0 + latSquare*1.0 + latSub*(2.5/60.0) + latExt*(15.0/3600.0) + (15.0/3600.0)/2.0
	}

	return Location{Grid: norm.String(), Lat: lat, Lon: lon}, nil
}
