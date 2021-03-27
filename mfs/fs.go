package mfs

import (
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func toErrno(err error) syscall.Errno {
	if e, ok := err.(syscall.Errno); ok {
		return e
	}
	return syscall.ENOENT
}

var getInode = func() func(path string) uint64 {
	var num uint64 = 10
	link := map[string]uint64{}
	return func(path string) uint64 {
		if _, ok := link[path]; !ok {
			link[path] = num
			num += 1
		}
		return link[path]
	}
}()

type Server struct {
	Realpath    string
	Transformer Transformer
	Options     []string
	Debug       bool
}

func (s *Server) Mount(mountPath string) error {
	dir := &Dir{
		Realpath: s.Realpath, Fakepath: mountPath, Transformer: s.Transformer}
	server, err := fs.Mount(mountPath, dir, &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug:   s.Debug,
			Options: s.Options,
		},
	})
	if err != nil {
		return err
	}
	server.Wait()
	return nil
}
