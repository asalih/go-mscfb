package mscfb

import (
	"fmt"
	"path"
	"strings"
	"unicode/utf16"
)

const MAX_NAME_LEN int = 31

type Ordering int

const (
	OrderLess Ordering = iota
	OrderEqual
	OrderGreater
)

func ValidateName(name string, nameUtf16 []uint16) error {
	if strings.ContainsAny(name, "/\\:!") {
		return fmt.Errorf("name contains one of /\\:! characters: %v", name)
	}

	return nil
}

func CompareNames(nameLeft, nameRight string) Ordering {
	nl := len(utf16.Encode([]rune(nameLeft)))
	nr := len(utf16.Encode([]rune(nameRight)))

	if nl == nr {
		if strings.EqualFold(nameLeft, nameRight) {
			return OrderEqual
		}
	}

	if nl > nr {
		return OrderGreater
	}

	return OrderLess
}

func NameChainFromPath(s string) []string {
	s = path.Clean(s)
	if s == "" {
		return []string{}
	}

	if s[0] == '/' {
		s = s[1:]
	}

	if s == "" {
		return []string{}
	}

	if strings.HasPrefix(s, "..") {
		return []string{}
	}

	return strings.Split(s, "/")
}

func PathFromNameChain(names []string) string {
	return "/" + strings.Join(names, "/")
}
