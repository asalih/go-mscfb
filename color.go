package mscfb

type Color int

const (
	Red Color = iota
	Black
)

func (c Color) AsByte() byte {
	switch c {
	case Red:
		return COLOR_RED
	case Black:
		return COLOR_BLACK
	default:
		return 0
	}
}

func ColorFromByte(b byte) Color {
	switch b {
	case COLOR_RED:
		return Red
	case COLOR_BLACK:
		return Black
	default:
		return -1
	}
}
