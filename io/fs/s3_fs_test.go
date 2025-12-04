package fs

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestNewS3FS(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY") == "" {
		t.Log("AWS_ACCESS_KEY is not set, skipping test")
		return
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
		Region:      aws.String("ap-northeast-1"),
	})
	require.NoError(t, err)
	protocol := s3.New(sess)
	s3FS := NewS3FS(protocol, "gridex-backend")

	nameFunc := func(name string) string {
		return "test/" + name
	}

	text := `
		Hello, world!
`
	err = s3FS.Upload(context.TODO(), nameFunc("test.txt"), strings.NewReader(text))
	require.NoError(t, err)

	ok, err := s3FS.Exists(context.TODO(), nameFunc("test.txt"))
	require.NoError(t, err)
	require.True(t, ok)

	// delete file
	defer func() {
		err := s3FS.Delete(context.TODO(), nameFunc("test.txt"))
		require.NoError(t, err)

		ok, err := s3FS.Exists(context.TODO(), nameFunc("test.txt"))
		require.NoError(t, err)
		require.False(t, ok)
	}()

	buf := bytes.NewBuffer(nil)
	err = s3FS.Download(context.TODO(), nameFunc("test.txt"), buf)
	require.NoError(t, err)
	require.Equal(t, text, buf.String())
}
