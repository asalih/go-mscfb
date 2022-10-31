package mscfb

type ObjectType int

const (
	ObjUnallocated ObjectType = iota
	ObjStorage
	ObjStream
	ObjRoot
)

func (o ObjectType) AsByte() byte {
	switch o {
	case ObjUnallocated:
		return OBJ_TYPE_UNALLOCATED
	case ObjStorage:
		return OBJ_TYPE_STORAGE
	case ObjStream:
		return OBJ_TYPE_STREAM
	case ObjRoot:
		return OBJ_TYPE_ROOT
	default:
		return 0
	}
}

func ObjectFromByte(b byte) ObjectType {
	switch b {
	case OBJ_TYPE_UNALLOCATED:
		return ObjUnallocated
	case OBJ_TYPE_STORAGE:
		return ObjStorage
	case OBJ_TYPE_STREAM:
		return ObjStream
	case OBJ_TYPE_ROOT:
		return ObjRoot
	default:
		return -1
	}
}
