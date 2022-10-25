package mscfb

import "fmt"

type Allocator struct {
	Sector         *Sector
	DifatSectorIds []uint32
	Difat          []uint32
	Fat            []uint32
	Validation     Validation
}

func NewAllocator(sector *Sector, difatSectorIds []uint32, difat []uint32, fat []uint32, validation Validation) (*Allocator, error) {
	alloc := Allocator{
		Sector:         sector,
		DifatSectorIds: difatSectorIds,
		Difat:          difat,
		Fat:            fat,
		Validation:     validation,
	}

	err := alloc.Validate()
	if err != nil {
		return nil, err
	}

	return &alloc, nil
}

func (a *Allocator) Next(index uint32) (uint32, error) {
	if index > uint32(len(a.Fat)) {
		return 0, fmt.Errorf("invalid index: %v", index)
	}

	nextId := a.Fat[index]
	if nextId != END_OF_CHAIN && (nextId > MAX_REGULAR_SECTOR || nextId >= uint32(len(a.Fat))) {
		return 0, fmt.Errorf("invalid next index: %v", nextId)
	}

	return nextId, nil
}

func (a *Allocator) Validate() error {
	if len(a.Fat) > int(a.Sector.NumSectors) {
		return fmt.Errorf("fat has %v entries, but file has %v: %w",
			len(a.Fat), a.Sector.NumSectors, ErrorInvalidCFB)
	}

	for _, difatSector := range a.DifatSectorIds {
		if difatSector >= uint32(len(a.Fat)) {
			return fmt.Errorf("invalid FAT has %v entries, but DIFAT lists %v as a DIFAT sector: %w",
				len(a.Fat), difatSector, ErrorInvalidCFB)
		}

		if a.Fat[difatSector] != DIFAT_SECTOR {
			if a.Validation.IsStrict() {
				return fmt.Errorf("invalid DIFAT sector %v is not marked as such in the FAT: %w", difatSector, ErrorInvalidCFB)
			} else {
				a.Fat[difatSector] = DIFAT_SECTOR
			}
		}
	}

	for _, difatSector := range a.Difat {
		if difatSector >= uint32(len(a.Fat)) {
			return fmt.Errorf("invalid FAT has %v entries, but DIFAT lists %v as a FAT sector: %w",
				len(a.Fat), difatSector, ErrorInvalidCFB)
		}

		if a.Fat[difatSector] != FAT_SECTOR {
			if a.Validation.IsStrict() {
				return fmt.Errorf("invalid FAT sector %v is not marked as such in the FAT: %w", difatSector, ErrorInvalidCFB)
			} else {
				a.Fat[difatSector] = FAT_SECTOR
			}
		}
	}

	pointees := make(map[uint32]bool)
	for fatIdx, fat := range a.Fat {
		if fat <= MAX_REGULAR_SECTOR {
			if fat >= uint32(len(a.Fat)) {
				return fmt.Errorf("invalid FAT entry %v points to sector %v, but file has only %v sectors: %w",
					fatIdx, fat, len(a.Fat), ErrorInvalidCFB)
			}
			if pointees[fat] {
				return fmt.Errorf("invalid FAT entry %v points to sector %v, which is already pointed to by another FAT entry: %w",
					fatIdx, fat, ErrorInvalidCFB)
			}
			pointees[fat] = true
		} else if fat == INVALID_SECTOR {
			return fmt.Errorf("invalid FAT entry %v points to sector %v, which is an invalid sector: %w", fatIdx, fat, ErrorInvalidCFB)
		}
	}

	return nil
}

func (a *Allocator) SeekToSector(sectorId uint32) (int64, error) {
	return a.Sector.SeekToSector(sectorId)
}

func (a *Allocator) SeekWithinSector(sectorId uint32, offset int64) (int64, error) {
	return a.Sector.SeekWithinSector(sectorId, offset)
}
