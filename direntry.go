package mscfb

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"

	"github.com/google/uuid"
)

type DirEntry struct {
	Name           string
	ObjType        ObjectType
	Color          Color
	LeftSibling    uint32
	RightSibling   uint32
	Child          uint32
	CLSID          uuid.UUID
	StateBits      uint32
	CreationTime   uint64
	ModifiedTime   uint64
	StartingSector uint32
	StreamSize     uint64
}

func NewDirEntry(name string, objType ObjectType, timestamp uint64) *DirEntry {
	dir := DirEntry{
		Name:         name,
		ObjType:      objType,
		Color:        Black,
		LeftSibling:  NO_STREAM,
		RightSibling: NO_STREAM,
		Child:        NO_STREAM,
		CLSID:        [16]byte{},
		StateBits:    0,
		CreationTime: timestamp,
		ModifiedTime: timestamp,
		StreamSize:   0,
	}
	if objType == ObjStorage {
		dir.StartingSector = 0
	} else {
		dir.StartingSector = END_OF_CHAIN
	}

	return &dir
}

func ReadDirEntry(reader io.ReadSeeker, version Version, validation Validation) (*DirEntry, error) {

	name := make([]uint16, 32)
	err := binary.Read(reader, binary.LittleEndian, &name)
	if err != nil {
		return nil, err
	}

	var nameLength uint16
	err = binary.Read(reader, binary.LittleEndian, &nameLength)
	if err != nil {
		return nil, err
	}

	if nameLength > 64 {
		return nil, fmt.Errorf("name length is too long: %v", nameLength)
	}
	if nameLength%2 != 0 {
		return nil, fmt.Errorf("name length is not even: %v", nameLength)
	}

	var nameCharLength uint16
	if nameLength > 0 {
		nameCharLength = (nameLength / 2) - 1
	}

	if validation.IsStrict() && name[nameCharLength] != 0 {
		return nil, fmt.Errorf("name is not null terminated")
	}

	nameStr := string(utf16.Decode(name[:nameCharLength]))

	var objTypeByte uint8
	err = binary.Read(reader, binary.LittleEndian, &objTypeByte)
	if err != nil {
		return nil, err
	}

	objType := ObjectFromByte(objTypeByte)
	if objType == -1 {
		return nil, fmt.Errorf("invalid object type: %v", objTypeByte)
	}

	// According to section 2.6.2 of the MS-CFB spec, "The root directory
	// entry's Name field MUST contain the null-terminated string 'Root
	// Entry' in Unicode UTF-16."  However, some CFB files in the wild
	// don't do this, so under Permissive validation we don't enforce it;
	// instead, for the root entry we just ignore the actual name in the
	// file and treat it as though it were what it's supposed to be.
	if objType == ObjRoot {
		if nameStr != ROOT_DIR_NAME && validation.IsStrict() {
			return nil, fmt.Errorf("root directory name is invalid: %v", nameStr)
		}
	}

	err = ValidateName(nameStr, name)
	if err != nil {
		return nil, err
	}

	var colorByte uint8
	err = binary.Read(reader, binary.LittleEndian, &colorByte)
	if err != nil {
		return nil, err
	}

	color := ColorFromByte(colorByte)
	if color == -1 {
		return nil, fmt.Errorf("invalid color: %v", colorByte)
	}

	var leftSibling uint32
	err = binary.Read(reader, binary.LittleEndian, &leftSibling)
	if err != nil {
		return nil, err
	}
	if leftSibling != NO_STREAM && leftSibling > MAX_REGULAR_SECTOR {
		return nil, fmt.Errorf("invalid left sibling: %v", leftSibling)
	}

	var rightSibling uint32
	err = binary.Read(reader, binary.LittleEndian, &rightSibling)
	if err != nil {
		return nil, err
	}
	if rightSibling != NO_STREAM && rightSibling > MAX_REGULAR_SECTOR {
		return nil, fmt.Errorf("invalid left sibling: %v", rightSibling)
	}

	var child uint32
	err = binary.Read(reader, binary.LittleEndian, &child)
	if err != nil {
		return nil, err
	}
	if child != NO_STREAM {
		if objType == ObjStream {
			return nil, fmt.Errorf("non-empty stream child: %v", child)
		}
		if child > MAX_REGULAR_SECTOR {
			return nil, fmt.Errorf("invalid child: %v", child)
		}
	}

	// Section 2.6.1 of the MS-CFB spec states that "In a stream object,
	// this [CLSID] field MUST be set to all zeroes."  However, some CFB
	// files in the wild violate this, so under Permissive validation we
	// don't enforce it; instead, for non-storage objects we just ignore
	// the CLSID data entirely and treat it as though it were nil.
	clsid, _ := readUuid(reader)
	if objType == ObjStream && clsid != uuid.Nil {
		if validation.IsStrict() {
			return nil, fmt.Errorf("non-nil CLSID for stream: %v", clsid)
		}
		clsid = uuid.Nil
	}

	var stateBits uint32
	err = binary.Read(reader, binary.LittleEndian, &stateBits)
	if err != nil {
		return nil, err
	}

	var creationTime uint64
	err = binary.Read(reader, binary.LittleEndian, &creationTime)
	if err != nil {
		return nil, err
	}

	var modifiedTime uint64
	err = binary.Read(reader, binary.LittleEndian, &modifiedTime)
	if err != nil {
		return nil, err
	}

	var startingSector uint32
	err = binary.Read(reader, binary.LittleEndian, &startingSector)
	if err != nil {
		return nil, err
	}

	var streamSize uint64
	err = binary.Read(reader, binary.LittleEndian, &streamSize)
	if err != nil {
		return nil, err
	}

	streamSize = streamSize & version.SectorLenMask()
	if objType == ObjStorage {
		if validation.IsStrict() && startingSector != 0 {
			return nil, fmt.Errorf("non-zero starting sector for storage: %v", startingSector)
		}
		startingSector = 0

		if validation.IsStrict() && streamSize != 0 {
			return nil, fmt.Errorf("non-zero stream size for storage: %v", streamSize)
		}
		streamSize = 0
	}

	dir := DirEntry{
		Name:           nameStr,
		ObjType:        objType,
		Color:          color,
		LeftSibling:    leftSibling,
		RightSibling:   rightSibling,
		Child:          child,
		CLSID:          clsid,
		StateBits:      stateBits,
		CreationTime:   creationTime,
		ModifiedTime:   modifiedTime,
		StartingSector: startingSector,
		StreamSize:     streamSize,
	}

	return &dir, nil
}

func readUuid(reader io.Reader) (uuid.UUID, error) {
	var d1 uint32
	var d2 uint16
	var d3 uint16
	var d4 [8]byte

	err := binary.Read(reader, binary.LittleEndian, &d1)
	if err != nil {
		return [16]byte{}, err
	}

	err = binary.Read(reader, binary.LittleEndian, &d2)
	if err != nil {
		return [16]byte{}, err
	}

	err = binary.Read(reader, binary.LittleEndian, &d3)
	if err != nil {
		return [16]byte{}, err
	}

	err = binary.Read(reader, binary.LittleEndian, &d4)
	if err != nil {
		return [16]byte{}, err
	}

	uuidBytes := make([]byte, 16)
	uuidBytes[0] = byte(d1 >> 24)
	uuidBytes[1] = byte(d1 >> 16)
	uuidBytes[2] = byte(d1 >> 8)
	uuidBytes[3] = byte(d1)
	uuidBytes[4] = byte(d2 >> 8)
	uuidBytes[5] = byte(d2)
	uuidBytes[6] = byte(d3 >> 8)
	uuidBytes[7] = byte(d3)
	copy(uuidBytes[8:], d4[:])

	return uuid.FromBytes(uuidBytes)
}
