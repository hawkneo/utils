package fs

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	systemFS "io/fs"
)

var _ FS = (*s3FS)(nil)

type s3FS struct {
	protocol *s3.S3
	bucket   string
}

func NewS3FS(protocol *s3.S3, bucket string) FS {
	return &s3FS{
		protocol: protocol,
		bucket:   bucket,
	}
}

func (fs *s3FS) Open(name string) (systemFS.File, error) {
	return fs.OpenWithContext(context.Background(), name)
}

func (fs *s3FS) OpenWithContext(ctx context.Context, name string) (systemFS.File, error) {
	output, err := fs.protocol.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(fs.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	f := &file{
		fileInfo: &fileInfo{
			name:    name,
			size:    *output.ContentLength,
			mode:    0600,
			modTime: *output.LastModified,
			isDir:   false,
			sys:     output.Metadata,
		},
		readFunc: func(bytes []byte) (int, error) {
			return output.Body.Read(bytes)
		},
		closeFunc: func() error {
			return output.Body.Close()
		},
	}
	return f, nil
}

func (fs *s3FS) ReadFile(name string) ([]byte, error) {
	return fs.ReadFileWithContext(context.Background(), name)
}

func (fs *s3FS) ReadFileWithContext(ctx context.Context, name string) ([]byte, error) {
	output, err := fs.protocol.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(fs.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()
	return io.ReadAll(output.Body)
}

func (fs *s3FS) ReadDir(name string) (entries []systemFS.DirEntry, err error) {
	return fs.ReadDirWithContext(context.Background(), name)
}

func (fs *s3FS) ReadDirWithContext(ctx context.Context, name string) (entries []systemFS.DirEntry, err error) {
	var continuationToken *string
	for {
		listObjects, err := fs.protocol.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(fs.bucket),
			ContinuationToken: continuationToken,
			Prefix:            aws.String(name),
		})
		if err != nil {
			return nil, err
		}

		for _, obj := range listObjects.Contents {
			entries = append(entries, &dirEntry{
				name:  *obj.Key,
				isDir: false,
				ftype: 0600,
				fileInfo: &fileInfo{
					name:    *obj.Key,
					size:    *obj.Size,
					mode:    0600,
					modTime: *obj.LastModified,
					isDir:   false,
					sys:     obj,
				},
			})
		}

		if *listObjects.IsTruncated == false {
			break
		}
		continuationToken = listObjects.NextContinuationToken
	}

	return entries, nil
}

func (fs *s3FS) Exists(ctx context.Context, name string) (bool, error) {
	output, err := fs.protocol.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(fs.bucket),
		Key:    aws.String(name),
	})
	defer func() {
		if output != nil && output.Body != nil {
			output.Body.Close()
		}
	}()

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == s3.ErrCodeNoSuchKey {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (fs *s3FS) Upload(ctx context.Context, name string, src io.ReadSeeker) error {
	_, err := fs.protocol.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:   src,
		Bucket: aws.String(fs.bucket),
		Key:    aws.String(name),
	})
	return err
}

func (fs *s3FS) Download(ctx context.Context, name string, dst io.Writer) error {
	file, err := fs.OpenWithContext(ctx, name)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(dst, file)
	return err
}

func (fs *s3FS) Delete(ctx context.Context, name string) error {
	_, err := fs.protocol.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(fs.bucket),
		Key:    aws.String(name),
	})
	return err
}
