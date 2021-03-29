package mfs

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type Dir struct {
	fs.Inode

	Realpath    string
	Fakepath    string
	Transformer Transformer
	Server      *Server

	cache map[string]string
}

var _ = (fs.NodeReaddirer)((*Dir)(nil))

var _ = (fs.NodeLookuper)((*Dir)(nil))

func (d *Dir) putCache(fake, real string) {
	if d.cache == nil {
		d.cache = make(map[string]string)
	}
	d.cache[fake] = real
}

func (d *Dir) getCache(fake string) (string, bool) {
	if d.cache == nil {
		d.cache = make(map[string]string)
	}
	real, ok := d.cache[fake]
	return real, ok
}

func (d *Dir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entries, err := os.ReadDir(d.Realpath)
	if err != nil {
		return nil, toErrno(err)
	}
	ret := make([]fuse.DirEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			return nil, toErrno(err)
		}
		out, err := d.Transformer.AttrTransform(&OpCtx{
			Context:  ctx,
			realpath: d.Realpath,
			fakepath: d.Fakepath,
			basepath: d.Server.Realpath,
		}, info)
		if err != nil {
			return nil, toErrno(err)
		}
		d.putCache(out.Name(), e.Name())
		ret = append(ret, fuse.DirEntry{
			Name: out.Name(),
			Mode: syscallModeFromFs(out.Mode()),
			Ino:  getInode(filepath.Join(d.Realpath, out.Name())),
		})
	}
	return fs.NewListDirStream(ret), 0
}

func (d *Dir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	realname, ok := d.getCache(name)
	if !ok {
		d.Readdir(ctx)
	}
	if realname, ok = d.getCache(name); !ok {
		return nil, syscall.ENOENT
	}
	realpath := filepath.Join(d.Realpath, realname)
	info, err := os.Lstat(realpath)
	if err != nil {
		return nil, toErrno(err)
	}
	finfo, err := d.Transformer.AttrTransform(&OpCtx{
		Context:  ctx,
		realpath: d.Realpath,
		fakepath: d.Fakepath,
		basepath: d.Server.Realpath,
	}, info)
	if err != nil {
		return nil, toErrno(err)
	}

	if finfo.Name() != name {
		return nil, syscall.ENOENT
	}
	fakepath := filepath.Join(d.Fakepath, finfo.Name())
	var inode fs.InodeEmbedder
	if finfo.Mode().IsDir() {
		inode = &Dir{
			Realpath:    realpath,
			Fakepath:    fakepath,
			Transformer: d.Transformer,
			Server:      d.Server,
		}
	} else {
		inode = &File{
			Realpath:    realpath,
			Fakepath:    fakepath,
			Transformer: d.Transformer,
			Server:      d.Server,
		}
	}
	return d.NewInode(ctx, inode, fs.StableAttr{
		Mode: syscallModeFromFs(finfo.Mode()),
		Ino:  getInode(realpath),
	}), 0
}
