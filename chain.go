package mscfb

import (
	"fmt"
	"io"
)

type Chain struct {
	Allocator       *Allocator
	SectorInit      SectorInit
	SectorIds       []uint32
	OffsetFromStart uint64
}

func NewChain(allocator *Allocator, startingSectorId uint32, init SectorInit) (*Chain, error) {
	sectorIds := make([]uint32, 0)
	currentSectorId := startingSectorId

	var err error
	for currentSectorId != END_OF_CHAIN {
		sectorIds = append(sectorIds, currentSectorId)
		currentSectorId, err = allocator.Next(currentSectorId)
		if err != nil {
			return nil, err
		}

		if currentSectorId == startingSectorId {
			return nil, fmt.Errorf("chain contained duplicate sector id %v", currentSectorId)
		}
	}

	return &Chain{
		Allocator:       allocator,
		SectorInit:      init,
		SectorIds:       sectorIds,
		OffsetFromStart: 0,
	}, nil
}

func (c *Chain) NumSectors() uint32 {
	return uint32(len(c.SectorIds))
}

func (c *Chain) Len() uint64 {
	return uint64(c.Allocator.Sectors.SectorLen() * len(c.SectorIds))
}

func (c *Chain) Read(p []byte) (int, error) {
	totalLen := c.Len()
	remainingInChain := totalLen - c.OffsetFromStart
	maxLen := min(uint64(len(p)), remainingInChain)
	if maxLen == 0 {
		return 0, io.EOF
	}

	sectorLen := uint64(c.Allocator.Sectors.SectorLen())
	currentSectorIndex := uint32(c.OffsetFromStart / sectorLen)
	currentSectorId := c.SectorIds[currentSectorIndex]
	offsetWithinSector := c.OffsetFromStart % sectorLen

	sector, err := c.Allocator.SeekWithinSector(currentSectorId, int64(offsetWithinSector))
	if err != nil {
		return 0, err
	}

	bytesReaded, err := sector.reader.Read(p[:maxLen])
	if err != nil {
		return 0, err
	}

	c.OffsetFromStart += uint64(bytesReaded)
	return bytesReaded, nil
}

func (c *Chain) Seek(offset int64, whence int) (int64, error) {
	length := c.Len()
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = int64(c.OffsetFromStart) + offset
	case io.SeekEnd:
		newOffset = int64(length) + offset
	}

	if newOffset < 0 || newOffset > int64(length) {
		return 0, fmt.Errorf("invalid offset %v", newOffset)
	}

	c.OffsetFromStart = uint64(newOffset)
	return int64(c.OffsetFromStart), nil
}

func (c *Chain) IntoSubSector(subSectorIndex uint32, subSectorLen int64, offsetWithin uint64) (*Sector, error) {
	subSectorPerSector := int64(c.Allocator.Sectors.SectorLen()) / subSectorLen
	sectorIndexWithinChain := subSectorIndex / uint32(subSectorPerSector)
	subsectorIndexWithinSector := subSectorIndex % uint32(subSectorPerSector)
	sectorId := c.SectorIds[sectorIndexWithinChain]

	sector, err := c.Allocator.SeekWithinSubSector(sectorId, subsectorIndexWithinSector, subSectorLen, int64(offsetWithin))
	if err != nil {
		return nil, err
	}

	return sector, nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
