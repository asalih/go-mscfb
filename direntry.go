package mscfb

import (
	"encoding/binary"
	"io"
)

type DirEntry struct {
	Name           string
	ObjType        ObjectType
	Color          Color
	LeftSibling    uint32
	RightSibling   uint32
	Child          uint32
	CLSID          [16]byte
	StateBits      uint32
	CreationTime   uint64
	ModifiedTime   uint64
	StartingSector uint32
	StreamSize     uint32
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
	if objType == Storage {
		dir.StartingSector = 0
	} else {
		dir.StartingSector = END_OF_CHAIN
	}

	return &dir
}

func ReadDirEntry(reader io.ReadSeeker, version Version, validation Validation) (*DirEntry, error) {
	dir := DirEntry{}

	if err := binary.Read(reader, binary.LittleEndian, &dir); err != nil {
		return nil, err
	}

	return &dir, nil
}
