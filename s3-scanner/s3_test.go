package tufin

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/require"
)

const (
	key  = "access-logs/10Jul2020/1.log"
	data = "This is S3 file mock data"
)

type mockS3Client struct {
	s3iface.S3API
}

func (m *mockS3Client) ListObjects(_ *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {

	return &s3.ListObjectsOutput{
		Contents: []*s3.Object{{Key: aws.String(key)}},
	}, nil
}

func TestS3Scanner_Scan(t *testing.T) {

	S3Scanner{
		svc:              &mockS3Client{},
		downloader:       getDownloader(),
		tmpDirFilesystem: ".",
	}.Scan("my-bucket", func(file *os.File) {
		defer func() { require.NoError(t, file.Close()) }()
		path := file.Name()
		require.Equal(t, key[strings.LastIndex(key, "/")+1:],
			path[strings.LastIndex(path, "/")+1:])
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		require.Equal(t, data, scanner.Text())
	})
}

func getDownloader() *s3manager.Downloader {

	var locker sync.Mutex
	payload := []byte(data)

	svc := s3.New(unit.Session)
	svc.Handlers.Send.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		locker.Lock()
		defer locker.Unlock()

		r.HTTPResponse = &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(payload)),
			Header:     http.Header{},
		}
		r.HTTPResponse.Header.Set("Content-Length", "1")
	})

	return s3manager.NewDownloaderWithClient(svc, func(d *s3manager.Downloader) {
		d.Concurrency = 1
		d.PartSize = 1
	})
}
