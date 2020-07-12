# Golang AWS S3 Scanner

Say you need to analyze files on AWS S3 bucket, 
mean you want to temporary download each bucket's file, 
do some work and delete the temporary file from your filesystem.
That is exactly what `S3Scanner` does, and it is *testable* :)

First, let's take a look on how the client looks like:
```go
tufin.NewScanner("us-east-2", ".").Scan("my-bucket", func(file *os.File) {
		log.Info(file.Name())
})
```
The _constructor_ gets AWS region and directory on client filesystem for creating temporary downloaded S3 files. 
_Scan_ gets a bucket to scan and foreach file it will send a pointer to a downloaded file.

### S3Scanner Implementation
Before getting list of file's names in a bucket, we need to create an SDK S3 service in `S3Scanner` constructor:
```go
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
```
The trick in here is to have a member with type `s3iface.S3API` for S3 service 
and not `*S3` like the `s3.New` function returns, so it will be testable.

Note, that when you initialize a new service client without supplying any arguments,
the AWS SDK attempts to find AWS credentials by using the default credential provider chain.
In our case `verifyEnv` verifies that `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` has non-empty values:
```go
func verifyEnv(keys ...string) {

	for _, currKey := range keys {
		if val := os.Getenv(currKey); val == "" {
			log.Fatalf("Please, set '%s'", val)
		}
	}
}
```

Now, we can call `ListObjects` using the `svc` we created in the constructor in `Scan` and easily mock it while testing:
```go
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
```
Last but not least, we're looping on S3 bucket files, download and call the client `func` on each downloaded file.
Let's take a look on how to download file from S3, which is done by the `*s3manager.Downloader` we created in the constructor:
```go
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
```

### Testing the S3Scanner
Let's start with mocking S3 `ListObjects`:
```go
type mockS3Client struct {
	s3iface.S3API
}

func (m *mockS3Client) ListObjects(_ *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {

	return &s3.ListObjectsOutput{
		Contents: []*s3.Object{{Key: aws.String(key)}},
	}, nil
}
```
As you can see it creates 1 Object with `key` which represents the file path on S3.

Mocking the `Download` function of `s3manager` will override `svc.Handlers.Send.PushBack`,
by streaming mock S3 file data:
```go
func getMockS3ForDownloader() s3iface.S3API {

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

	return svc
}
```
Let's put all together:
```go
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

	downloader := s3manager.NewDownloaderWithClient(getMockS3ForDownloader(), func(d *s3manager.Downloader) {
		d.Concurrency = 1
		d.PartSize = 1
	})

	S3Scanner{
		svc:              mockS3Client{},
		downloader:       downloader,
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

func getMockS3ForDownloader() s3iface.S3API {

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

	return svc
}
```

See a full example in [GitHub](See a full example: https://github.com/Tufin/blog/tree/master/s3-scanner)

Reference: [Go SDK S3 Docs](https://docs.aws.amazon.com/sdk-for-go/api/service/s3/)
