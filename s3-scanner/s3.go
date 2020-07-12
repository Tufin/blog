package tufin

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
)

type S3Scanner struct {
	svc              s3iface.S3API
	downloader       *s3manager.Downloader
	tmpDirFilesystem string
}

// NewS3Scanner creates s3 service and downloader
func NewS3Scanner(region string, dir string) *S3Scanner {

	// verify aws auth
	verifyEnv("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY")

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.Fatalf("failed to create aws session to region '%s' with '%v'", region, err)
	}
	s3Svc := s3.New(awsSession)

	return &S3Scanner{
		svc:              s3Svc,
		downloader:       s3manager.NewDownloaderWithClient(s3Svc),
		tmpDirFilesystem: dir}
}

// Iterate on S3 bucket's files - download it into temporary directory
// and sends 'touch' event on each file
func (fi S3Scanner) Scan(bucket string, touch func(*os.File)) {

	response, err := fi.svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Fatalf("failed to get list of s3 files from bucket '%s' with '%v'", bucket, err)
	}

	for _, currObj := range response.Contents {
		currS3FilePath := *currObj.Key
		currS3FileName := currS3FilePath[strings.LastIndex(currS3FilePath, "/")+1:]
		currFilePath := fmt.Sprintf("%s/%s", fi.tmpDirFilesystem, currS3FileName)
		touch(download(fi.downloader, bucket, currObj.Key, currS3FilePath, currFilePath))
		deleteFromFilesystem(currFilePath)
	}
}

func verifyEnv(keys ...string) {

	for _, currKey := range keys {
		if val := os.Getenv(currKey); val == "" {
			log.Fatalf("Please, set '%s'", val)
		}
	}
}

func download(downloader *s3manager.Downloader, bucket string, key *string, s3FilePath string, path string) *os.File {

	log.Debugf("Creating a file '%s'...", path)
	ret, err := os.Create(path)
	if err != nil {
		log.Fatalf("failed to create file '%s'", path)
	}

	log.Infof("Downloading file '%s' from s3...", path)
	_, err = downloader.Download(ret, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    key,
	})
	if err != nil {
		log.Fatalf("unable to download '%s' from bucket '%s' with '%v'", s3FilePath, bucket, err)
	}

	return ret
}

func deleteFromFilesystem(path string) {

	if err := os.Remove(path); err != nil {
		log.Errorf("failed to delete file '%s' from filesystem with '%v'", path, err)
	}
}
