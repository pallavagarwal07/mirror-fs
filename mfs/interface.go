package mfs

import (
	"context"
	"io/fs"
	"os"
)

type Ctx interface {
	context.Context

	Realpath() string
	Fakepath() string
}

type OpCtx struct {
	context.Context

	realpath string
	fakepath string
}

func (o *OpCtx) Realpath() string { return o.realpath }
func (o *OpCtx) Fakepath() string { return o.fakepath }

var _ Ctx = (*OpCtx)(nil)

type FileInfo interface {
	Name() string
	Mode() fs.FileMode
}

type AttrTransformer interface {
	AttrTransform(Ctx, os.FileInfo) (FileInfo, error)
}

type DataTransformer interface {
	DataTransform(Ctx, []byte) ([]byte, error)
}

type Transformer interface {
	AttrTransformer
	DataTransformer
}

type ReverseTransformer interface {
	ReverseTransform(Ctx, []byte) ([]byte, error)
}

type clone struct{}

var Clone *clone = &clone{}

func (*clone) AttrTransform(ctx Ctx, info os.FileInfo) (FileInfo, error) {
	return info, nil
}

func (*clone) DataTransform(ctx Ctx, input []byte) ([]byte, error) {
	return input, nil
}

func (*clone) ReverseTransform(ctx Ctx, input []byte) ([]byte, error) {
	return input, nil
}
