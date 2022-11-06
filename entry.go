package mscfb

import (
	"path"

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

func (e *Entry) IsStream() bool {
	return e.ObjType == ObjStream
}

func (e *Entry) IsStorage() bool {
	return e.ObjType == ObjStorage || e.ObjType == ObjRoot
}

func (e *Entry) IsRoot() bool {
	return e.ObjType == ObjRoot
}

type EntriesOrder int

const (
	EntriesNonRecursive EntriesOrder = iota
	EntriesPreorder
)

type EntriesStack struct {
	ParentPath    string
	StreamId      uint32
	VisitSiblings bool
}

type Entries struct {
	Order     EntriesOrder
	Directory *Directory
	Stack     []*EntriesStack
}

func NewEntries(order EntriesOrder, directory *Directory, parentPath string, start uint32) *Entries {
	entries := &Entries{
		Order:     order,
		Directory: directory,
		Stack:     make([]*EntriesStack, 0),
	}

	if order == EntriesNonRecursive {
		entries.StackLeftSpine(parentPath, start)
	} else if order == EntriesPreorder {
		entries.Stack = append(entries.Stack, &EntriesStack{
			ParentPath:    parentPath,
			StreamId:      start,
			VisitSiblings: false,
		})
	}

	return entries
}

func (e *Entries) StackLeftSpine(parentPath string, currentId uint32) {
	for currentId != NO_STREAM {
		currentEntry := e.Directory.DirEntries[currentId]

		e.Stack = append(e.Stack, &EntriesStack{
			ParentPath:    parentPath,
			StreamId:      currentId,
			VisitSiblings: true,
		})

		currentId = currentEntry.LeftSibling
	}
}

func (e *Entries) Next() *Entry {
	if len(e.Stack) == 0 {
		return nil
	}

	//pop stack
	currentStack := e.Stack[len(e.Stack)-1]
	e.Stack = e.Stack[:len(e.Stack)-1]

	dirEntry := e.Directory.DirEntries[currentStack.StreamId]
	path := joinPath(currentStack.ParentPath, dirEntry)
	if currentStack.VisitSiblings {
		e.StackLeftSpine(currentStack.ParentPath, dirEntry.RightSibling)
	}

	if e.Order == EntriesPreorder &&
		dirEntry.ObjType == ObjStream &&
		dirEntry.Child != NO_STREAM {
		e.StackLeftSpine(path, dirEntry.Child)
	}

	return NewEntry(dirEntry, path)
}

func joinPath(parentPath string, dirEntry *DirEntry) string {
	if dirEntry.ObjType == ObjRoot {
		return parentPath
	}

	return path.Join(parentPath, dirEntry.Name)
}
