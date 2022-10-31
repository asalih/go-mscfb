package mscfb

import (
	"github.com/google/uuid"
)

type Entry struct {
	Name         string
	Path         string
	ObjType      ObjectType
	CLSID        uuid.UUID
	StateBits    uint32
	CreationTime uint64
	ModifiedTime uint64
	StreamLen    uint64
}

func NewEntry(dirEntry *DirEntry, path string) *Entry {
	entry := Entry{
		Name:         dirEntry.Name,
		Path:         path,
		ObjType:      dirEntry.ObjType,
		CLSID:        dirEntry.CLSID,
		StateBits:    dirEntry.StateBits,
		CreationTime: dirEntry.CreationTime,
		ModifiedTime: dirEntry.ModifiedTime,
		StreamLen:    dirEntry.StreamSize,
	}

	return &entry
}
