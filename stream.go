package mscfb

import (
	"io"
)

const BUFFER_SIZE uint32 = 8192

type Stream struct {
	CompoundFile *CompoundFile

	StreamId        uint32
	TotalLen        uint64
	Buffer          []byte
	Position        uint64
	Cap             uint64
	OffsetFromStart uint64
}

func newStream(comp *CompoundFile, streamId uint32) *Stream {
	totalLen := comp.MiniAlloc.Directory.DirEntries[streamId].StreamSize
	return &Stream{
		CompoundFile: comp,

		StreamId:        streamId,
		TotalLen:        totalLen,
		Buffer:          make([]byte, BUFFER_SIZE),
		Position:        0,
		Cap:             0,
		OffsetFromStart: 0,
	}
}

func (s *Stream) CurrentPosition() uint64 {
	return s.OffsetFromStart + uint64(s.Position)
}

func (s *Stream) Read(p []byte) (int, error) {
	if s.Position >= s.Cap &&
		s.CurrentPosition() < s.TotalLen {
		s.OffsetFromStart += uint64(s.Position)
		s.Position = 0

		cap, err := s.readDataFromStream()
		if err != nil {
			return 0, err
		}

		s.Cap = uint64(cap)
	}

	if s.CurrentPosition() == s.Cap {
		return 0, io.EOF
	}

	numBytes := copy(p, s.Buffer[s.Position:s.Cap])

	s.Position = min(s.Cap, s.Position+uint64(numBytes))

	return numBytes, nil
}

func (s *Stream) readDataFromStream() (int, error) {
	dirEntry := s.CompoundFile.MiniAlloc.Directory.DirEntries[s.StreamId]

	var numBytes int
	if s.OffsetFromStart >= dirEntry.StreamSize {
		numBytes = 0
	} else {
		remaining := dirEntry.StreamSize - s.OffsetFromStart
		if remaining < uint64(len(s.Buffer)) {
			numBytes = int(remaining)
		} else {
			numBytes = len(s.Buffer)
		}
	}

	if numBytes > 0 {
		if dirEntry.StreamSize < uint64(MINI_STREAM_CUTOFF) {
			chain, err := s.CompoundFile.MiniAlloc.OpenMiniChain(dirEntry.StartingSector)
			if err != nil {
				return 0, err
			}

			_, err = chain.Seek(int64(s.OffsetFromStart), io.SeekStart)
			if err != nil {
				return 0, err
			}

			_, err = chain.ReadAll(s.Buffer[:numBytes])
			if err != nil {
				return 0, err
			}
		} else {
			chain, err := s.CompoundFile.Directory.Allocator.OpenChain(dirEntry.StartingSector, SectorInitZero)
			if err != nil {
				return 0, err
			}

			_, err = chain.Seek(int64(s.OffsetFromStart), io.SeekStart)
			if err != nil {
				return 0, err
			}

			_, err = chain.ReadAll(s.Buffer[:numBytes])
			if err != nil {
				return 0, err
			}
		}
	}

	return numBytes, nil
}

func (s *Stream) Seek(offset int64, whence int) (int64, error) {

	var pos int64
	switch whence {
	case io.SeekStart:
		pos = offset
		if pos > int64(s.TotalLen) {
			return 0, io.EOF
		}
	case io.SeekCurrent:
		cpos := s.CurrentPosition()
		pos = int64(cpos) + offset
		if cpos > s.TotalLen {
			return 0, io.EOF
		}
		if cpos < 0 {
			return 0, io.EOF
		}

	case io.SeekEnd:
		pos = int64(s.TotalLen) + offset
		if pos > 0 {
			return 0, io.EOF
		}
	}

	if pos < int64(s.OffsetFromStart) || pos > int64(s.OffsetFromStart+s.Cap) {
		s.OffsetFromStart = uint64(pos)
		s.Position = 0
		s.Cap = 0
	} else {
		s.Position = uint64(pos - int64(s.OffsetFromStart))
	}

	return int64(s.Position), nil

}
