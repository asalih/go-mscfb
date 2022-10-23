package mscfb

type ObjectType int

const (
	Unallocated ObjectType = iota
	Storage
	Stream
	Root
)

func (o ObjectType) AsByte() byte {
	switch o {
	case Unallocated:
		return OBJ_TYPE_UNALLOCATED
	case Storage:
		return OBJ_TYPE_STORAGE
	case Stream:
		return OBJ_TYPE_STREAM
	case Root:
		return OBJ_TYPE_ROOT
	default:
		return 0
	}
}

func ObjectFromByte(b byte) ObjectType {
	switch b {
	case OBJ_TYPE_UNALLOCATED:
		return Unallocated
	case OBJ_TYPE_STORAGE:
		return Storage
	case OBJ_TYPE_STREAM:
		return Stream
	case OBJ_TYPE_ROOT:
		return Root
	default:
		return Unallocated
	}
}
