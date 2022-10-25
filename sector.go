package mscfb

import (
	"fmt"
	"io"
)

type SectorInit int

const (
	SectorInitZero SectorInit = iota
	SectorInitFat
	SectorInitDifat
	SectorInitDir
)

func (s SectorInit) Initialize(sector *Sector) {
	panic("not implemented")
}

type Sector struct {
	Version    Version
	NumSectors uint32

	reader io.ReadSeeker
}

func NewSector(v Version, bufferLength int64, reader io.ReadSeeker) *Sector {
	sectorLen := v.SectorLen()
	numSectors := ((bufferLength + int64(sectorLen) - 1) / int64(sectorLen)) - 1

	return &Sector{
		Version:    v,
		NumSectors: uint32(numSectors),
		reader:     reader,
	}
}

func (s *Sector) SectorLen() int {
	return s.Version.SectorLen()
}

func (s *Sector) SeekToSector(sectorId uint32) (int64, error) {
	return s.SeekWithinSector(sectorId, 0)
}

func (s *Sector) SeekWithinSector(sectorId uint32, offset int64) (int64, error) {
	if sectorId >= s.NumSectors {
		return 0, fmt.Errorf("tried to seek to sector %v, but sector count is only %v", sectorId, s.NumSectors)
	}

	return s.reader.Seek((int64(sectorId)+1)*int64(s.SectorLen())+offset, io.SeekStart)
}
