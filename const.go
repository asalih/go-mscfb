package mscfb

// ========================================================================= //

const (
	HEADER_LEN                  int = 512 // length of CFB file header, in bytes
	DIR_ENTRY_LEN               int = 128 // length of directory entry, in bytes
	NUM_DIFAT_ENTRIES_IN_HEADER int = 109
)

// Constants for CFB file header values:
var MAGIC_NUMBER = []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}

const (
	MINOR_VERSION      int    = 0x3e
	BYTE_ORDER_MARK    uint16 = 0xfffe
	MINI_SECTOR_SHIFT  uint16 = 6 // 64-byte mini sectors
	MINI_SECTOR_LEN    int    = 1 << (MINI_SECTOR_SHIFT)
	MINI_STREAM_CUTOFF uint32 = 4096
)

// Constants for FAT entries:
const (
	MAX_REGULAR_SECTOR uint32 = 0xfffffffa
	INVALID_SECTOR     uint32 = 0xfffffffb
	DIFAT_SECTOR       uint32 = 0xfffffffc
	FAT_SECTOR         uint32 = 0xfffffffd
	END_OF_CHAIN       uint32 = 0xfffffffe
	FREE_SECTOR        uint32 = 0xffffffff
)

// Constants for directory entries:
const (
	ROOT_DIR_NAME                = "Root Entry"
	OBJ_TYPE_UNALLOCATED  uint8  = 0
	OBJ_TYPE_STORAGE      uint8  = 1
	OBJ_TYPE_STREAM       uint8  = 2
	OBJ_TYPE_ROOT         uint8  = 5
	COLOR_RED             uint8  = 0
	COLOR_BLACK           uint8  = 1
	ROOT_STREAM_ID        uint32 = 0
	MAX_REGULAR_STREAM_ID uint32 = 0xfffffffa
	NO_STREAM             uint32 = 0xffffffff
)

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
