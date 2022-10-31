package mscfb

import "fmt"

type MiniAlloc struct {
	Directory          *Directory
	Minifat            []uint32
	MinifatStartSector uint32
}

func NewMiniAlloc(d *Directory, minifat []uint32, minifatStartSector uint32) (*MiniAlloc, error) {
	alloc := MiniAlloc{
		Directory:          d,
		Minifat:            minifat,
		MinifatStartSector: minifatStartSector,
	}

	err := alloc.Validate()
	if err != nil {
		return nil, err
	}

	return &alloc, nil
}

func (a *MiniAlloc) Validate() error {
	rootEntry := a.Directory.RootDirEntry()
	rootStreamMiniSectors := rootEntry.StreamSize / uint64(MINI_SECTOR_LEN)
	if rootStreamMiniSectors < uint64(len(a.Minifat)) {
		return fmt.Errorf("miniFAT has %v entries, but root stream has only %v mini sectors",
			len(a.Minifat), rootStreamMiniSectors)
	}

	pointees := make(map[uint32]bool)
	for miniSectorIdx, miniSector := range a.Minifat {
		if miniSector <= MAX_REGULAR_SECTOR {
			if miniSector >= uint32(len(a.Minifat)) {
				return fmt.Errorf("miniFAT[%v] points to mini sector %v, but there are only %v mini sectors",
					miniSectorIdx, miniSector, len(a.Minifat))
			}

			if pointees[miniSector] {
				return fmt.Errorf("mini sector %v pointed to twice", miniSector)
			}

			pointees[miniSector] = true
		}
	}

	return nil
}

func (a *MiniAlloc) RootDirEntry() *Entry {
	return NewEntry(a.Directory.RootDirEntry(), "/")
}

func (a *MiniAlloc) StreamIDForNameChain(names []string) (uint32, error) {
	return a.Directory.StreamIDForNameChain(names)
}

func (a *MiniAlloc) OpenMiniChain(sectorId uint32) (*MiniChain, error) {
	return NewMiniChain(a, sectorId)
}

func (a *MiniAlloc) Next(sectorId uint32) (uint32, error) {
	if sectorId >= uint32(len(a.Minifat)) {
		return 0, fmt.Errorf("sector id %v out of range", sectorId)
	}

	nextId := a.Minifat[sectorId]
	if nextId != END_OF_CHAIN &&
		(nextId > MAX_REGULAR_SECTOR ||
			nextId >= uint32(len(a.Minifat))) {
		return 0, fmt.Errorf("sector id %v points to invalid sector id %v", sectorId, nextId)
	}

	return nextId, nil
}

func (a *MiniAlloc) SeekWithinMiniSector(sectorId uint32, offset uint64) (*Sector, error) {
	miniStreamStartSector := a.Directory.RootDirEntry().StartingSector
	chain, err := a.Directory.Allocator.OpenChain(miniStreamStartSector, SectorInitFat)
	if err != nil {
		return nil, err
	}

	return chain.IntoSubSector(sectorId, int64(MINI_SECTOR_LEN), offset)
}
