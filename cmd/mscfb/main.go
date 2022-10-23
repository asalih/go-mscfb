package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/asalih/go-mscfb"
)

func main() {
	files, err := filepath.Glob("testdata/*.msi")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file == "testdata/.DS_Store" {
			continue
		}

		rdr, err := os.OpenFile(file, os.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}

		msi, err := mscfb.Open(rdr, mscfb.ValidationPermissive)
		if err != nil {
			panic(err)
		}

		fmt.Println(msi)
	}
}
