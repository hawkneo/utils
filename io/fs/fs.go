package fs

import (
	"context"
	"io"
	systemFS "io/fs"
	"time"
)

type FS interface {
	systemFS.FS
	OpenWithContext(ctx context.Context, name string) (systemFS.File, error)

	systemFS.ReadFileFS
	ReadFileWithContext(ctx context.Context, name string) ([]byte, error)

	systemFS.ReadDirFS
	ReadDirWithContext(ctx context.Context, name string) ([]systemFS.DirEntry, error)

	Exists(ctx context.Context, name string) (bool, error)

	Upload(ctx context.Context, name string, src io.ReadSeeker) error

	Download(ctx context.Context, name string, dst io.Writer) error

	Delete(ctx context.Context, name string) error
}

var _ systemFS.DirEntry = (*dirEntry)(nil)

type dirEntry struct {
	name     string
	isDir    bool
	ftype    systemFS.FileMode
	fileInfo *fileInfo
}

func (d *dirEntry) Name() string {
	return d.name
}

func (d *dirEntry) IsDir() bool {
	return d.isDir
}

func (d *dirEntry) Type() systemFS.FileMode {
	return d.ftype
}

func (d *dirEntry) Info() (systemFS.FileInfo, error) {
	return d.fileInfo, nil
}

var _ systemFS.File = (*file)(nil)

type file struct {
	fileInfo  *fileInfo
	readFunc  func([]byte) (int, error)
	closeFunc func() error
}

func (f *file) Stat() (systemFS.FileInfo, error) {
	return f.fileInfo, nil
}

func (f *file) Read(bytes []byte) (int, error) {
	return f.readFunc(bytes)
}

func (f *file) Close() error {
	return f.closeFunc()
}

var _ systemFS.FileInfo = (*fileInfo)(nil)

type fileInfo struct {
	name    string
	size    int64
	mode    systemFS.FileMode
	modTime time.Time
	isDir   bool
	sys     any
}

func (info *fileInfo) Name() string {
	return info.name
}

func (info *fileInfo) Size() int64 {
	return info.size
}

func (info *fileInfo) Mode() systemFS.FileMode {
	return info.mode
}

func (info *fileInfo) ModTime() time.Time {
	return info.modTime
}

func (info *fileInfo) IsDir() bool {
	return info.isDir
}

func (info *fileInfo) Sys() any {
	return info.sys
}
