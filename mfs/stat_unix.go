package mfs

import (
	"io/fs"
	"syscall"
)

func syscallModeFromFs(mode fs.FileMode) uint32 {
	ret := uint32(mode) & 0777
	switch true {
	case mode&fs.ModeDevice&fs.ModeCharDevice > 0:
		ret |= syscall.S_IFCHR
	case mode&fs.ModeDevice > 0:
		ret |= syscall.S_IFBLK
	case mode&fs.ModeDir > 0:
		ret |= syscall.S_IFDIR
	case mode&fs.ModeNamedPipe > 0:
		ret |= syscall.S_IFIFO
	case mode&fs.ModeSymlink > 0:
		ret |= syscall.S_IFLNK
	case mode&fs.ModeSocket > 0:
		ret |= syscall.S_IFSOCK
	default:
		ret |= syscall.S_IFREG
	}

	if mode&fs.ModeSetgid > 0 {
		mode |= syscall.S_ISGID
	}
	if mode&fs.ModeSetuid > 0 {
		mode |= syscall.S_ISUID
	}
	if mode&fs.ModeSticky > 0 {
		mode |= syscall.S_ISVTX
	}
	return ret
}
