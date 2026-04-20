package geocode

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	citiesBinVersion uint32 = 1
)

var citiesBinMagic = [4]byte{'M', 'M', 'C', 'B'}

// City is a population center with its administrative hierarchy codes.
type City struct {
	Name        string
	Lat         float32
	Lon         float32
	CountryCode string // 2-char ISO
	Admin1Code  string // up to 8 chars
	Admin2Code  string // up to 8 chars
}

type cityRecord struct {
	Lat         float32
	Lon         float32
	CountryCode [2]byte
	Admin1Code  [8]byte
	Admin2Code  [8]byte
	NameLen     uint16
}

// WriteCitiesBin encodes cities to the MMCB v1 binary format.
func WriteCitiesBin(w io.Writer, cities []City) error {
	if err := binary.Write(w, binary.LittleEndian, citiesBinMagic); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, citiesBinVersion); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(cities))); err != nil {
		return err
	}

	for _, c := range cities {
		name := []byte(c.Name)
		if len(name) > 0xFFFF {
			return fmt.Errorf("city name too long (%d bytes): %q", len(name), c.Name)
		}
		rec := cityRecord{
			Lat:     c.Lat,
			Lon:     c.Lon,
			NameLen: uint16(len(name)),
		}
		copyFixed(rec.CountryCode[:], c.CountryCode)
		copyFixed(rec.Admin1Code[:], c.Admin1Code)
		copyFixed(rec.Admin2Code[:], c.Admin2Code)

		if err := binary.Write(w, binary.LittleEndian, rec); err != nil {
			return err
		}
		if _, err := w.Write(name); err != nil {
			return err
		}
	}
	return nil
}

// ReadCitiesBin decodes cities from the MMCB v1 binary format.
func ReadCitiesBin(r io.Reader) ([]City, error) {
	var magic [4]byte
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return nil, fmt.Errorf("read magic: %w", err)
	}
	if magic != citiesBinMagic {
		return nil, fmt.Errorf("bad magic: got %v, want MMCB", magic)
	}
	var version, count uint32
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}
	if version != citiesBinVersion {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("read count: %w", err)
	}

	cities := make([]City, 0, count)
	for i := uint32(0); i < count; i++ {
		var rec cityRecord
		if err := binary.Read(r, binary.LittleEndian, &rec); err != nil {
			return nil, fmt.Errorf("read record %d: %w", i, err)
		}
		nameBuf := make([]byte, rec.NameLen)
		if _, err := io.ReadFull(r, nameBuf); err != nil {
			return nil, fmt.Errorf("read name %d: %w", i, err)
		}
		cities = append(cities, City{
			Name:        string(nameBuf),
			Lat:         rec.Lat,
			Lon:         rec.Lon,
			CountryCode: trimFixed(rec.CountryCode[:]),
			Admin1Code:  trimFixed(rec.Admin1Code[:]),
			Admin2Code:  trimFixed(rec.Admin2Code[:]),
		})
	}
	return cities, nil
}

func copyFixed(dst []byte, s string) {
	n := copy(dst, s)
	for i := n; i < len(dst); i++ {
		dst[i] = 0
	}
}

func trimFixed(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
