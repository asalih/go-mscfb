package mscfb

import (
	"fmt"
	"io"
)

type MiniChain struct {
	MiniAlloc *MiniAlloc
	SectorIds []uint32
	Offset    uint64
}

func NewMiniChain(miniAlloc *MiniAlloc, sectorId uint32) (*MiniChain, error) {
	sectorIds := make([]uint32, 0)
	currentSectorId := sectorId
	firstSectorId := sectorId

	var err error
	for currentSectorId != END_OF_CHAIN {
		sectorIds = append(sectorIds, currentSectorId)
		currentSectorId, err = miniAlloc.Next(currentSectorId)
		if err != nil {
			return nil, err
		}

		if currentSectorId == firstSectorId {
			return nil, fmt.Errorf("chain contained duplicate sector id %v", currentSectorId)
		}
	}

	return &MiniChain{
		MiniAlloc: miniAlloc,
		SectorIds: sectorIds,
		Offset:    0,
	}, nil
}

func (c *MiniChain) Len() uint64 {
	return uint64(MINI_SECTOR_LEN * len(c.SectorIds))
}

func (c *MiniChain) Read(p []byte) (n int, err error) {
	totalLen := c.Len()
	remainingInChain := totalLen - c.Offset
	maxLen := min(uint64(len(p)), remainingInChain)
	if maxLen == 0 {
		return 0, io.EOF
	}

	sectorLen := uint64(MINI_SECTOR_LEN)
	currentSectorIndex := uint32(c.Offset / sectorLen)
	currentSectorId := c.SectorIds[currentSectorIndex]
	offsetWithinSector := c.Offset % sectorLen

	sector, err := c.MiniAlloc.SeekWithinMiniSector(currentSectorId, offsetWithinSector)
	if err != nil {
		return 0, err
	}

	bytesRead, err := sector.Read(p)
	if err != nil {
		return 0, err
	}

	c.Offset += uint64(bytesRead)

	return bytesRead, nil
}

func (c *MiniChain) ReadAll(p []byte) (int, error) {
	shouldRead := cap(p)
	totalRead := 0

	for {
		remainig := shouldRead - totalRead
		if remainig == 0 {
			return totalRead, nil
		}

		n, err := c.Read(p[totalRead:])
		totalRead += n

		if err == io.EOF {
			return totalRead, nil
		}

		if err != nil {
			return totalRead, err
		}
	}
}

func (c *MiniChain) Seek(offset int64, whence int) (int64, error) {
	length := c.Len()
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = int64(c.Offset) + offset
	case io.SeekEnd:
		newOffset = int64(length) + offset
	}

	if newOffset < 0 || newOffset > int64(length) {
		return 0, fmt.Errorf("invalid offset %v", newOffset)
	}

	c.Offset = uint64(newOffset)
	return int64(c.Offset), nil
}
