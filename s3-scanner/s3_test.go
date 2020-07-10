package tufin

import (
	"bufio"
	"bytes"
	"fmt"
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

type mockS3Client struct {
	s3iface.S3API
}

const (
	key  = "access-logs/10Jul202/1.log"
	data = "This is mock S3 file data"
)

func (m *mockS3Client) ListObjects(*s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {

	return &s3.ListObjectsOutput{
		Contents: []*s3.Object{{Key: aws.String(key)}},
	}, nil
}

func TestS3FileIterator_Iterate(t *testing.T) {

	downloader := s3manager.NewDownloaderWithClient(getMockS3ForDownloader(), func(d *s3manager.Downloader) {
		d.Concurrency = 1
		d.PartSize = 1
	})

	S3Scanner{
		svc:              &mockS3Client{},
		downloader:       downloader,
		tmpDirFilesystem: ".",
	}.Scan("my-bucket", func(file *os.File) {
		defer require.NoError(t, file.Close())
		path := file.Name()
		require.Equal(t, key[strings.LastIndex(key, "/")+1:],
			path[strings.LastIndex(path, "/")+1:])
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		require.Equal(t, data, scanner.Text())
	})
}

func getMockS3ForDownloader() s3iface.S3API {

	var locker sync.Mutex
	payload := []byte(data)

	svc := s3.New(unit.Session)
	svc.Handlers.Send.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		locker.Lock()
		defer locker.Unlock()

		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(payload[:])),
			Header:     http.Header{},
		}
		r.HTTPResponse.Header.Set("Content-Length", fmt.Sprintf("%d", len(payload)))
	})

	return svc
}
