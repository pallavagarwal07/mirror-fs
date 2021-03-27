package mfs

import (
	"context"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type File struct {
	fs.Inode

	Realpath    string
	Fakepath    string
	Transformer Transformer
}

type FileHandle struct {
	file *File

	data []byte
}

var _ = (fs.NodeOpener)((*File)(nil))
var _ = (fs.NodeSetattrer)((*File)(nil))

var _ = (fs.FileReader)((*FileHandle)(nil))
var _ = (fs.FileGetattrer)((*FileHandle)(nil))

var _ = (fs.FileWriter)((*FileHandle)(nil))
var _ = (fs.FileSetattrer)((*FileHandle)(nil))

func slice(data []byte, start, end int64) []byte {
	size := int64(len(data))
	if start >= size || start >= end {
		return []byte{}
	}
	if end >= size {
		end = size
	}
	return data[start:end]
}

func (f *File) Open(
	ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	data, err := os.ReadFile(f.Realpath)
	if err != nil {
		return nil, 0, toErrno(err)
	}
	data, err = f.Transformer.DataTransform(
		&OpCtx{Context: ctx, realpath: f.Realpath, fakepath: f.Fakepath}, data)
	if err != nil {
		return nil, 0, toErrno(err)
	}
	return &FileHandle{file: f, data: data}, 0, 0
}

func (f *FileHandle) Read(
	_ context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	return fuse.ReadResultData(slice(f.data, off, off+int64(len(dest)))), 0
}

func (fh *FileHandle) Getattr(
	ctx context.Context, out *fuse.AttrOut) syscall.Errno {
	out.Size = uint64(len(fh.data))
	out.SetTimes(nil, nil, nil)
	out.Mode = 0440
	return 0
}

func (fh *FileHandle) Setattr(
	ctx context.Context, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	if _, ok := fh.file.Transformer.(ReverseTransformer); !ok {
		return syscall.EROFS
	}
	backup := make([]byte, len(fh.data))
	copy(fh.data, backup)

	if sz, ok := in.GetSize(); ok {
		if sz > uint64(len(backup)) {
			backup = append(backup, make([]byte, sz-uint64(len(backup)))...)
		}
		backup = backup[:sz]
	}
	if err := fh.write(ctx, backup); err != 0 {
		return err
	}
	return fh.Getattr(ctx, out)
}

func (f *File) Setattr(
	ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	if fh, ok := fh.(*FileHandle); ok {
		return fh.Setattr(ctx, in, out)
	}
	fh, _, err := f.Open(ctx, 0)
	if err != 0 {
		return err
	}
	return fh.(*FileHandle).Setattr(ctx, in, out)
}

func (fh *FileHandle) write(ctx context.Context, backup []byte) syscall.Errno {
	f := fh.file
	out, err := f.Transformer.(ReverseTransformer).ReverseTransform(
		&OpCtx{Context: ctx, realpath: f.Realpath, fakepath: f.Fakepath}, backup)
	if err != nil {
		return toErrno(err)
	}
	if err := ioutil.WriteFile(f.Realpath, out, 0644); err != nil {
		return toErrno(err)
	}
	fh.data = backup
	return 0
}

func (fh *FileHandle) Write(
	ctx context.Context, data []byte, o int64) (uint32, syscall.Errno) {
	off, f := int(o), fh.file
	if _, ok := f.Transformer.(ReverseTransformer); !ok {
		return 0, syscall.EROFS
	}

	backup := make([]byte, len(fh.data))
	copy(fh.data, backup)

	if off > len(backup) {
		backup = append(backup, make([]byte, len(backup)-off)...)
	}
	var i int
	for i = 0; i < len(data); i++ {
		if off+i >= len(backup) {
			break
		}
		backup[off+i] = data[i]
	}
	backup = append(backup, data[i:]...)

	if err := fh.write(ctx, backup); err != 0 {
		return 0, err
	}
	return uint32(len(data)), 0
}
