package main

import (
	"log"
	"os"

	"github.com/pallavagarwal07/mirror-fs/mfs"
)

func main() {
	// This is where we'll mount the FS
	srcDir := "/tmp/source"
	os.Mkdir(srcDir, 0755)
	dstDir := "/tmp/dest"
	os.Mkdir(dstDir, 0755)
	root := &mfs.Server{
		Realpath:    srcDir,
		Transformer: mfs.Clone,
		Debug:       true,
		Options:     []string{"allow_other", "default_permissions"},
	}
	if err := root.Mount(dstDir); err != nil {
		log.Fatalln(err)
	}
}
