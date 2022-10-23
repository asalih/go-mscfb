package mscfb

import (
	"errors"
	"io"
)

var (
	ErrorInvalidCFB = errors.New("invalid cfb file")
)

type CompoundFile struct {
	r io.ReadSeeker
}

func Open(reader io.ReadSeeker) (*CompoundFile, error) {
	_, err := reader.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	header := make([]byte, HEADER_LEN)
	len, err := reader.Read(header)
	if err != nil {
		return nil, err
	}

	if len < HEADER_LEN {
		return nil, ErrorInvalidCFB
	}

	compoundFile := CompoundFile{
		r: reader,
	}

	return &compoundFile, nil
}
