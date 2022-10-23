package mscfb

import "fmt"

const (
	V3 Version = 3
	V4 Version = 4
)

type Version int

func VersionNumber(v uint16) (Version, error) {
	switch v {
	case 3:
		return V3, nil
	case 4:
		return V4, nil
	default:
		return 0, fmt.Errorf("invalid version number: %v", v)
	}
}

// Returns the sector shift used in this version.
func (v Version) SectorShift() uint16 {
	return uint16(v * 3)
}

// Returns the length of sectors used in this version.
func (v Version) SectorLen() int {
	return 1 << v.SectorShift()
}

// Returns the bitmask used for reading stream lengths in this version.
func (v Version) SectorLenMask() uint64 {
	switch v {
	case V3:
		return 0xffffffff
	case V4:
		return 0xffffffffffffffff
	default:
		return 0
	}
}

// Returns the number of directory entries per sector in this version.
func (v Version) DirEntriesPerSector() int {
	return v.SectorLen() / DIR_ENTRY_LEN
}
