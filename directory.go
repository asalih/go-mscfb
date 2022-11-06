package mscfb

import "fmt"

type Directory struct {
	Allocator      *Allocator
	DirEntries     []*DirEntry
	DirStartSector uint32
}

func NewDirectory(allocator *Allocator, dirEntries []*DirEntry, dirStartSector uint32) (*Directory, error) {
	dir := Directory{
		Allocator:      allocator,
		DirEntries:     dirEntries,
		DirStartSector: dirStartSector,
	}

	err := dir.Validate()
	if err != nil {
		return nil, err
	}

	return &dir, nil
}

func (d *Directory) RootDirEntry() *DirEntry {
	return d.DirEntries[ROOT_STREAM_ID]
}

// Returns an iterator over the entries within the root storage object.
func (d *Directory) RootStorageEntries() *Entries {
	start := d.RootDirEntry().Child

	return NewEntries(EntriesNonRecursive, d, PathFromNameChain([]string{}), start)
}

func (d *Directory) Validate() error {
	if len(d.DirEntries) == 0 {
		return fmt.Errorf("directory has no entries")
	}

	rootDirEntry := d.RootDirEntry()
	if rootDirEntry == nil {
		return fmt.Errorf("directory has no root entry")
	}

	if rootDirEntry.StreamSize%uint64(MINI_SECTOR_LEN) != 0 {
		return fmt.Errorf("root stream len is %v, but should be multiple of %v", rootDirEntry.StreamSize, MINI_SECTOR_LEN)
	}

	visited := make(map[uint32]bool)
	stack := []uint32{ROOT_STREAM_ID}

	for len(stack) > 0 {
		dirEntryId := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[dirEntryId] {
			return fmt.Errorf("directory has a cycle")
		}

		visited[dirEntryId] = true

		dirEntry := d.DirEntries[dirEntryId]
		if dirEntry == nil {
			return fmt.Errorf("directory has no entry for id %v", dirEntryId)
		}

		if dirEntryId == ROOT_STREAM_ID {
			if dirEntry.ObjType != ObjRoot {
				return fmt.Errorf("root entry has object type: %v", dirEntry.ObjType)
			}
		} else if dirEntry.ObjType != ObjStorage && dirEntry.ObjType != ObjStream {
			return fmt.Errorf("non-root entry with object type: %v", dirEntry.ObjType)
		}

		leftSibling := dirEntry.LeftSibling
		if leftSibling != NO_STREAM {
			if leftSibling >= uint32(len(d.DirEntries)) {
				return fmt.Errorf("left sibling index is %v, but directory entry count is %v",
					leftSibling, len(d.DirEntries))
			}

			entry := d.DirEntries[leftSibling]
			if CompareNames(entry.Name, dirEntry.Name) != OrderLess {
				return fmt.Errorf("name ordering, %v vs %v", entry.Name, dirEntry.Name)
			}

			stack = append(stack, leftSibling)
		}

		rightSibling := dirEntry.RightSibling
		if rightSibling != NO_STREAM {
			if rightSibling >= uint32(len(d.DirEntries)) {
				return fmt.Errorf("right sibling index is %v, but directory entry count is %v",
					rightSibling, len(d.DirEntries))
			}

			entry := d.DirEntries[rightSibling]
			if CompareNames(dirEntry.Name, entry.Name) != OrderLess {
				return fmt.Errorf("name ordering, %v vs %v", entry.Name, dirEntry.Name)
			}

			stack = append(stack, rightSibling)
		}

		child := dirEntry.Child
		if child != NO_STREAM {
			if child >= uint32(len(d.DirEntries)) {
				return fmt.Errorf("child index is %v, but directory entry count is %v",
					child, len(d.DirEntries))
			}

			stack = append(stack, child)
		}
	}

	return nil
}

func (d *Directory) StreamIDForNameChain(names []string) (uint32, error) {
	streamId := ROOT_STREAM_ID

	for _, name := range names {
		streamId = d.DirEntries[streamId].Child
		for {
			if streamId == NO_STREAM {
				return 0, fmt.Errorf("stream not found: %v", name)
			}
			dirEntry := d.DirEntries[streamId]
			order := CompareNames(name, dirEntry.Name)
			if order == OrderEqual {
				break
			}

			switch order {
			case OrderLess:
				streamId = dirEntry.LeftSibling
			case OrderGreater:
				streamId = dirEntry.RightSibling
			}
		}
	}

	return streamId, nil
}
