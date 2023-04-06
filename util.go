package main

import "path/filepath"

func filenameWithoutExt(path string) string {
	fileName := filepath.Base(path)
	extension := filepath.Ext(fileName)
	name := fileName[0 : len(fileName)-len(extension)]

	return name
}
