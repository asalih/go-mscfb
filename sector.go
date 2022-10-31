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

func (s SectorInit) Initialize(sector *Sectors) {
	panic("not implemented")
}

type Sectors struct {
	Version    Version
	NumSectors uint32

	inner io.ReadSeeker
}

type Sector struct {
	SectorLen int64
	Offset    int64

	reader io.ReadSeeker
}

func NewSectors(v Version, bufferLength int64, reader io.ReadSeeker) *Sectors {
	sectorLen := v.SectorLen()
	numSectors := ((bufferLength + int64(sectorLen) - 1) / int64(sectorLen)) - 1

	return &Sectors{
		Version:    v,
		NumSectors: uint32(numSectors),
		inner:      reader,
	}
}

func (s *Sectors) SectorLen() int {
	return s.Version.SectorLen()
}

func (s *Sectors) SeekToSector(sectorId uint32) (*Sector, error) {
	return s.SeekWithinSector(sectorId, 0)
}

func (s *Sectors) SeekWithinSector(sectorId uint32, offset int64) (*Sector, error) {
	if sectorId >= s.NumSectors {
		return nil, fmt.Errorf("tried to seek to sector %v, but sector count is only %v", sectorId, s.NumSectors)
	}

	_, err := s.inner.Seek(int64(sectorId+1)*int64(s.SectorLen())+offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return &Sector{
		SectorLen: int64(s.SectorLen()),
		Offset:    offset,
		reader:    s.inner,
	}, nil
}

func (s *Sector) SubSector(start, len int64) (*Sector, error) {
	return &Sector{
		SectorLen: len,
		Offset:    s.Offset - start,
		reader:    s.reader,
	}, nil
}

func (s *Sector) Remaining() int64 {
	return s.SectorLen - s.Offset
}

func (s *Sector) Read(p []byte) (int, error) {
	maxLen := min(uint64(len(p)), uint64(s.Remaining()))
	if maxLen == 0 {
		return 0, io.EOF
	}

	bytesReaded, err := s.reader.Read(p[:maxLen])
	if err != nil {
		return 0, err
	}

	s.Offset += int64(bytesReaded)
	return bytesReaded, nil
}
