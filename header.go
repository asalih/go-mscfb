package mscfb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Header struct {
	Version            Version
	NumDirSectors      int
	NumFatSectors      int
	FirstDirSector     int
	FirstMinifatSector int
	NumMinifatSector   int
	FirstDifatSector   int
	NumDifatSectors    int

	InitialDifatEntries []int
}

const (
	reservedAfterMagicNumber = 16
	reservedAfterMiniShift   = 6
)

func (h *Header) readFrom(reader io.ReadSeeker) error {
	magicPart := make([]byte, len(MAGIC_NUMBER))
	_, err := reader.Read(magicPart)
	if err != nil {
		return err
	}

	if !bytes.Equal(magicPart, MAGIC_NUMBER) {
		return ErrorInvalidCFB
	}

	// seek reserved field
	_, err = reader.Seek(reservedAfterMagicNumber, io.SeekCurrent)
	if err != nil {
		return err
	}

	var minorVersion uint16
	err = binary.Read(reader, binary.LittleEndian, &minorVersion)
	if err != nil {
		return err
	}

	var versionNumber uint16
	err = binary.Read(reader, binary.LittleEndian, &versionNumber)
	if err != nil {
		return err
	}

	var byteOrderMark uint16
	err = binary.Read(reader, binary.LittleEndian, &byteOrderMark)
	if err != nil {
		return err
	}

	if byteOrderMark != BYTE_ORDER_MARK {
		return fmt.Errorf("Invalid CFB byte order mark (expected 0x{:04X}, found 0x{:04X})", BYTE_ORDER_MARK, byteOrderMark)
	}

	version, err := VersionNumber(versionNumber)
	if err != nil {
		return err
	}

	var sectorShift uint16
	err = binary.Read(reader, binary.LittleEndian, &sectorShift)
	if err != nil {
		return err
	}
	if sectorShift != version.SectorShift() {
		return fmt.Errorf("incorrect sector shift for CFB version %v (expected %v, found %v)", version, version.SectorShift(), sectorShift)
	}

	var miniSectorShift uint16
	err = binary.Read(reader, binary.LittleEndian, &miniSectorShift)
	if err != nil {
		return err
	}
	if miniSectorShift != MINI_SECTOR_SHIFT {
		return fmt.Errorf("incorrect mini sector shift (expected %v, found %v)", MINI_SECTOR_SHIFT, miniSectorShift)
	}

	// seek reserved field
	_, err = reader.Seek(reservedAfterMiniShift, io.SeekCurrent)
	if err != nil {
		return err
	}

	var numDirSectors uint32
	var numFatSectors uint32
	var firstDirSector uint32
	var transactionSign uint32

	err = binary.Read(reader, binary.LittleEndian, &numDirSectors)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.LittleEndian, &numFatSectors)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.LittleEndian, &firstDirSector)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.LittleEndian, &transactionSign)
	if err != nil {
		return err
	}

	var miniStreamCutoff uint32
	err = binary.Read(reader, binary.LittleEndian, &miniStreamCutoff)
	if err != nil {
		return err
	}
	if miniStreamCutoff != MINI_STREAM_CUTOFF {
		return fmt.Errorf("incorrect mini stream cutoff (expected %v, found %v)", MINI_STREAM_CUTOFF, miniStreamCutoff)
	}

	var firstMinifatSector uint32
	var numMinifatSectors uint32
	var firstDifatSector uint32
	var numDifatSectors uint32

	err = binary.Read(reader, binary.LittleEndian, &firstMinifatSector)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.LittleEndian, &numMinifatSectors)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.LittleEndian, &firstDifatSector)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.LittleEndian, &numDifatSectors)
	if err != nil {
		return err
	}

	// Some CFB implementations use FREE_SECTOR to indicate END_OF_CHAIN.
	if firstDifatSector == FREE_SECTOR {
		firstDifatSector = END_OF_CHAIN
	}

}
