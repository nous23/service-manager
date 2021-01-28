package global

import (
	"fmt"
	"os"
	"path/filepath"
)

func init() {
	d, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(fmt.Sprintf("get current dir failed: %v", err))
	}
	CurrDir = d
	StaticDir = filepath.Join(CurrDir, "static")
}

// current binary dir
var CurrDir string

// dir to store static web files
var StaticDir string
