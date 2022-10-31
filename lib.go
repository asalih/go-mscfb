package mscfb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

var (
	ErrorInvalidCFB = errors.New("invalid cfb file")
)

type CompoundFile struct {
	Reader io.ReadSeeker

	Header    *Header
	Directory *Directory
	MiniAlloc *MiniAlloc
}

func Open(reader io.ReadSeeker, validation Validation) (*CompoundFile, error) {
	bufLen, err := reader.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	if int(bufLen) < HEADER_LEN {
		return nil, ErrorInvalidCFB
	}

	_, err = reader.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	header := &Header{}
	err = header.readFrom(reader)
	if err != nil {
		return nil, err
	}

	sectorLen := header.Version.SectorLen()
	if bufLen > ((int64(MAX_REGULAR_SECTOR) + 1) * int64(sectorLen)) {
		return nil, fmt.Errorf("file is too large: %w", ErrorInvalidCFB)
	}

	if bufLen < int64(sectorLen) {
		return nil, fmt.Errorf("file is too small: %w", ErrorInvalidCFB)
	}

	sectors := NewSectors(header.Version, bufLen, reader)

	difat := make([]uint32, len(header.InitialDifatEntries))
	copy(difat, header.InitialDifatEntries)

	seenSectorIds := make(map[uint32]bool)
	difatSectorIds := make([]uint32, 0)
	currentDifatSector := header.FirstDifatSector

	var sz uint32
	uSize := unsafe.Sizeof(sz)

	for currentDifatSector != END_OF_CHAIN {
		if currentDifatSector > MAX_REGULAR_SECTOR {
			return nil, fmt.Errorf("invalid DIFAT chain: %w", ErrorInvalidCFB)
		} else if currentDifatSector >= sectors.NumSectors {
			return nil, fmt.Errorf("invalid DIFAT chain includes sector index: %w", ErrorInvalidCFB)
		}

		if seenSectorIds[currentDifatSector] {
			return nil, fmt.Errorf("DIFAT chain includes duplicate sector index: %w", ErrorInvalidCFB)
		}

		seenSectorIds[currentDifatSector] = true
		difatSectorIds = append(difatSectorIds, currentDifatSector)

		sector, err := sectors.SeekToSector(currentDifatSector)
		if err != nil {
			return nil, err
		}

		for i := 0; i < (sectors.SectorLen()/int(uSize) - 1); i++ {
			var next uint32
			err = binary.Read(sector.reader, binary.LittleEndian, &next)
			if err != nil {
				return nil, err
			}

			if next != FREE_SECTOR && next > MAX_REGULAR_SECTOR {
				return nil, fmt.Errorf("invalid DIFAT refers to invalid sector index %v", next)
			}
			difat = append(difat, next)
		}

		err = binary.Read(sector.reader, binary.LittleEndian, &currentDifatSector)
		if err != nil {
			return nil, err
		}
	}

	if validation.IsStrict() &&
		header.NumDifatSectors != uint32(len(difatSectorIds)) {
		return nil, fmt.Errorf("incorrect DIFAT chain length (header says %v, actual is %v): %w",
			header.NumDifatSectors, len(difatSectorIds), ErrorInvalidCFB)
	}

	//difat pop
	for i := len(difat) - 1; i >= 0; i-- {
		if difat[i] != FREE_SECTOR {
			break
		}
		difat = difat[:i]
	}

	if validation.IsStrict() &&
		header.NumFatSectors != uint32(len(difat)) {
		return nil, fmt.Errorf("incorrect number of FAT sectors (header says %v, DIFAT says %v)",
			header.NumFatSectors, len(difat))
	}

	fat := make([]uint32, 0)
	for _, sectorId := range difat {
		if sectorId >= sectors.NumSectors {
			return nil, fmt.Errorf("invalid FAT sector index: %w", ErrorInvalidCFB)
		}

		sector, err := sectors.SeekToSector(sectorId)
		if err != nil {
			return nil, err
		}
		for i := 0; i < sectors.SectorLen()/int(uSize); i++ {
			var next uint32
			err = binary.Read(sector.reader, binary.LittleEndian, &next)
			if err != nil {
				return nil, err
			}
			fat = append(fat, next)
		}
	}

	//fat pop
	if !validation.IsStrict() {
		for len(fat) > int(sectors.NumSectors) && fat[len(fat)-1] == 0 {
			fat = fat[:len(fat)-1]
		}
	}

	for i := len(fat) - 1; i >= 0; i-- {
		if fat[i] != FREE_SECTOR {
			break
		}
		fat = fat[:i]
	}

	allocator, err := NewAllocator(sectors, difatSectorIds, difat, fat, validation)
	if err != nil {
		return nil, err
	}

	// Read in directory.
	dirEntries := make([]*DirEntry, 0)
	seenDirSectors := make(map[uint32]bool)
	currentDirSector := header.FirstDirSector

	for currentDirSector != END_OF_CHAIN {
		if currentDirSector > MAX_REGULAR_SECTOR {
			return nil, fmt.Errorf("invalid directory chain: %w", ErrorInvalidCFB)
		} else if currentDirSector >= sectors.NumSectors {
			return nil, fmt.Errorf("invalid directory chain includes sector index: %w", ErrorInvalidCFB)
		}

		if seenDirSectors[currentDirSector] {
			return nil, fmt.Errorf("directory chain includes duplicate sector index: %w", ErrorInvalidCFB)
		}

		seenDirSectors[currentDirSector] = true

		_, err = allocator.SeekToSector(currentDirSector)
		if err != nil {
			return nil, err
		}

		for i := 0; i < header.Version.DirEntriesPerSector(); i++ {
			entry, err := ReadDirEntry(reader, header.Version, validation)
			if err != nil {
				return nil, err
			}

			dirEntries = append(dirEntries, entry)
		}

		currentDirSector, err = allocator.Next(currentDirSector)
		if err != nil {
			return nil, err
		}
	}

	directory, err := NewDirectory(allocator, dirEntries, header.FirstDirSector)
	if err != nil {
		return nil, err
	}

	chain, err := NewChain(allocator, header.FirstMinifatSector, SectorInitFat)
	if err != nil {
		return nil, err
	}

	if validation.IsStrict() && header.NumMinifatSector != chain.NumSectors() {
		return nil, fmt.Errorf("incorrect number of MiniFAT sectors (header says %v, FAT says %v)",
			header.NumMinifatSector, chain.NumSectors())
	}

	numMinifatEntries := uint32(chain.Len() / 4)
	minifat := make([]uint32, 0)

	p := []byte{0, 0, 0, 0}
	for i := uint32(0); i < numMinifatEntries; i++ {
		_, err := chain.Read(p)
		if err != nil {
			return nil, err
		}
		minifat = append(minifat, binary.LittleEndian.Uint32(p))
	}

	for i := len(minifat) - 1; i >= 0; i-- {
		if minifat[i] != FREE_SECTOR {
			break
		}
		minifat = minifat[:i]
	}

	miniAlloc, err := NewMiniAlloc(directory, minifat, header.FirstMinifatSector)
	if err != nil {
		return nil, err
	}

	compoundFile := CompoundFile{
		Reader: reader,

		Header:    header,
		Directory: directory,
		MiniAlloc: miniAlloc,
	}

	return &compoundFile, nil
}

func (c *CompoundFile) RootEntry() *Entry {
	return NewEntry(c.Directory.RootDirEntry(), "/")
}

func (c *CompoundFile) OpenStream(path string) (*Stream, error) {
	names := NameChainFromPath(path)
	path = PathFromNameChain(names)
	streamId, err := c.MiniAlloc.StreamIDForNameChain(names)
	if err != nil {
		return nil, err
	}

	if streamId == 0 {
		return nil, fmt.Errorf("stream not found: %s", path)
	}

	entry := c.MiniAlloc.Directory.DirEntries[streamId]
	if entry.ObjType != ObjStream {
		return nil, fmt.Errorf("not a stream: %s", path)
	}

	return newStream(c, streamId), nil
}
