# Golang AWS S3 Scanner

Say you need to analyze files on AWS S3 bucket, 
mean you want to temporary download each bucket's file, 
do some work and delete the temporary file from your filesystem.
That is exactly what `S3Scanner` does with `Scan` and it is *testable* :)

First, let's take a look on how the client looks like:
```golang
tufin.NewScanner("us-east-2", ".").Iterate("my-bucket", func(file *os.File) {
		log.Info(file.Name())
})
```
The _constructor_ gets AWS region and directory on client filesystem for creating temporary downloaded S3 files. 
_Iterate_ gets bucket to scan and foreach file it will send a pointer to a downloaded file.



