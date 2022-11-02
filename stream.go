package mscfb

import (
	"fmt"
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
	bufData, err := s.fillBuf()
	if err != nil {
		return 0, err
	}

	if s.CurrentPosition() >= s.TotalLen {
		return 0, io.EOF
	}

	numBytes := copy(p, bufData)

	s.Position = min(s.Cap, s.Position+uint64(numBytes))

	return numBytes, nil
}

func (s *Stream) fillBuf() ([]byte, error) {
	if s.Position >= s.Cap &&
		s.CurrentPosition() < s.TotalLen {
		s.OffsetFromStart += uint64(s.Position)
		s.Position = 0

		cap, err := s.readDataFromStream()
		if err != nil {
			return nil, err
		}

		s.Cap = uint64(cap)
	}

	return s.Buffer[s.Position:s.Cap], nil
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

func (s *Stream) Seek(pos int64, whence int) (int64, error) {
	delta := pos
	var newPos int64

	switch whence {
	case io.SeekStart:
		if delta > int64(s.TotalLen) {
			return 0, fmt.Errorf("cannot seek to %v bytes from start, because stream length is only %v bytes",
				delta, s.TotalLen)
		}
		newPos = delta

	case io.SeekCurrent:
		oldPos := s.CurrentPosition()
		if delta < 0 {
			delta = -delta
			if delta > int64(oldPos) {
				return 0, fmt.Errorf("cannot seek backwards %v bytes, because current position is only %v bytes",
					delta, oldPos)
			}
			newPos = int64(oldPos) - delta
		} else {
			remaining := s.TotalLen - oldPos
			if delta > int64(remaining) {
				return 0, fmt.Errorf("cannot seek forward %v bytes, because only %v bytes remain in stream",
					delta, remaining)
			}
			newPos = int64(oldPos) + delta
		}

	case io.SeekEnd:
		if delta > 0 {
			return 0, fmt.Errorf("cannot seek to %v bytes from end, because stream length is only %v bytes",
				delta, s.TotalLen)
		} else {
			delta = -delta
			if delta > int64(s.TotalLen) {
				return 0, fmt.Errorf("cannot seek to %v bytes from end, because stream length is only %v bytes",
					delta, s.TotalLen)
			}
		}
		newPos = int64(s.TotalLen) - delta
	}

	if newPos < int64(s.OffsetFromStart) || newPos > int64(s.OffsetFromStart+s.Cap) {
		s.OffsetFromStart = uint64(newPos)
		s.Position = 0
		s.Cap = 0
	} else {
		s.Position = uint64(newPos - int64(s.OffsetFromStart))
	}

	return newPos, nil
}
