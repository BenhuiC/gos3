package ceph

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"gos3/config"
	"gos3/pkg/util"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	B = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
	EB
)

type AwsClient struct {
	*s3.S3
	uploader *s3manager.Uploader
	*config.OSSConfig
	UploadPool *ants.Pool
}

var (
	Client *AwsClient
)

func InitAwsClient(cfg config.OSSConfig) error {
	var err error
	if Client, err = NewAwsClient(cfg); err != nil {
		return err
	}
	return nil
}

func NewAwsClient(cfg config.OSSConfig) (*AwsClient, error) {
	cli := AwsClient{
		OSSConfig: &cfg,
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials:               credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.AccessKeySecret, ""),
		Endpoint:                  aws.String(cfg.Endpoint),
		Region:                    aws.String("us-east-1"),
		DisableSSL:                aws.Bool(false),
		DisableEndpointHostPrefix: &cfg.DisableEndpointHostPrefix,
		S3ForcePathStyle:          &cfg.S3ForcePathStyle,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				// ResponseHeaderTimeout: time.Second * 5,
				Dial: (&net.Dialer{
					Timeout: 2 * time.Second,
				}).Dial,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.InsecureSkipVerify,
					MinVersion:         tls.VersionTLS12,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	cli.uploader = s3manager.NewUploader(sess, func(uploader *s3manager.Uploader) {
		uploader.Concurrency = 20
		uploader.PartSize = 10 * 1024 * 1024 // 10MB
	})
	cli.S3 = s3.New(sess)

	uploadPool, err := ants.NewPool(10)
	if err != nil {
		return nil, err
	}
	cli.UploadPool = uploadPool
	return &cli, nil
}

func (cli *AwsClient) RemoveFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}
	objs := []*s3.ObjectIdentifier{}
	for i := range files {
		objs = append(objs, &s3.ObjectIdentifier{
			Key: util.Point(files[i]),
		})
	}
	_, err := cli.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: &cli.Bucket,
		Delete: &s3.Delete{
			Objects: objs,
		},
	})
	return err
}

func (cli *AwsClient) GetObjectBytes(obj string) ([]byte, error) {
	out, err := cli.GetObject(obj)
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	bf := bytes.NewBuffer([]byte{})
	_, err = bf.ReadFrom(out.Body)
	return bf.Bytes(), err
}

func (cli *AwsClient) GetObjectToWriter(obj string, writer io.Writer) (err error) {
	var out *s3.GetObjectOutput
	out, err = cli.GetObject(obj)
	if err != nil {
		return
	}
	defer out.Body.Close()
	_, err = io.Copy(writer, out.Body)
	return
}

func (cli *AwsClient) GetObject(key string) (*s3.GetObjectOutput, error) {
	output, err := cli.S3.GetObject(&s3.GetObjectInput{
		Bucket: &cli.Bucket,
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (cli *AwsClient) DownloadObject(key, filename string) (size int64, err error) {
	output, err := cli.GetObject(key)
	if err != nil {
		return 0, err
	}
	defer output.Body.Close()

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	size, err = io.Copy(f, output.Body)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (cli *AwsClient) GetObjectSize(key string) (int64, error) {
	header, err := cli.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(cli.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, err
	}
	return *header.ContentLength, nil
}

func (cli *AwsClient) WithGetObject(obj string, fn func(out io.Reader) error) error {
	out, err := cli.GetObject(obj)
	if err != nil {
		return err
	}
	defer out.Body.Close()
	return fn(out.Body)
}

type PreSignedParam struct {
	Key            string
	Expires        time.Duration
	WantedFileName *string
}

func (cli *AwsClient) PreSignedGetURL(p PreSignedParam) (url string, err error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(cli.Bucket),
		Key:    aws.String(p.Key),
	}

	if p.WantedFileName != nil {
		input.ResponseContentDisposition = aws.String(fmt.Sprintf(`filename="%v"`, *p.WantedFileName))
	}
	req, _ := cli.S3.GetObjectRequest(input)
	url, err = req.Presign(p.Expires)
	return
}

func (cli *AwsClient) UploadFile(fileName string, key string, expire time.Duration) (size int64, err error) {
	var f *os.File
	if f, err = os.Open(fileName); err != nil {
		return
	}
	defer f.Close()
	var stat os.FileInfo
	stat, err = f.Stat()
	if err != nil {
		return
	}

	size = stat.Size()
	if stat.Size() > GB {
		params := &s3manager.UploadInput{
			Bucket: aws.String(cli.Bucket),
			Key:    aws.String(key),
			Body:   f,
		}
		if expire != 0 {
			params.Expires = aws.Time(time.Now().Add(expire))
		}
		_, err = cli.uploader.Upload(params)
	} else {
		params := &s3.PutObjectInput{
			Bucket: aws.String(cli.Bucket),
			Key:    aws.String(key),
			Body:   f,
		}
		if expire != 0 {
			params.Expires = aws.Time(time.Now().Add(expire))
		}
		_, err = cli.PutObject(params)
	}

	return
}

func (cli *AwsClient) UploadByReader(obj string, f io.Reader, expire time.Duration) (err error) {
	params := &s3manager.UploadInput{
		Bucket: aws.String(cli.Bucket),
		Key:    aws.String(obj),
		Body:   f,
	}
	if expire != 0 {
		params.Expires = aws.Time(time.Now().Add(expire))
	}
	_, err = cli.uploader.Upload(params)
	return
}

func (cli *AwsClient) CompleteMultipartUpload(uploadID, key string, completedPart []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	out, err := cli.S3.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(cli.Bucket),
		UploadId: aws.String(uploadID),
		Key:      aws.String(key),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedPart,
		},
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (cli *AwsClient) CreateMultipartUpload(obj string, expires time.Duration) (uploadID string, err error) {
	out, err := cli.S3.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:  &cli.Bucket,
		Key:     &obj,
		Expires: util.Point(time.Now().Add(expires)),
	})
	if err != nil {
		return
	}
	return *out.UploadId, err
}

var (
	ImagePreSignExpires = time.Minute
)

// no remote call
func (cli *AwsClient) GetImagePresignedURL(f string) (string, error) {
	return cli.PreSignedGetURL(PreSignedParam{
		Key:     f,
		Expires: ImagePreSignExpires,
		// WantedFileName: ,
	})
}

func (cli *AwsClient) PreSignedPutURL(key string, exp time.Duration) (string, error) {
	return cli.generatePreSignedURLImpl(key, "PUT", exp)
}

func (cli *AwsClient) PreSignUploadPartUrl(uploadID, key string, partNumber, contentLength int64, expiredIn time.Duration) (string, error) {
	req, _ := cli.UploadPartRequest(&s3.UploadPartInput{
		Bucket:        aws.String(cli.Bucket),
		UploadId:      aws.String(uploadID),
		ContentLength: aws.Int64(contentLength),
		PartNumber:    aws.Int64(partNumber),
		Key:           aws.String(key),
	})
	return cli.presign(req, expiredIn)
}

func (cli *AwsClient) generatePreSignedURLImpl(key string, method string, exp time.Duration) (string, error) {
	var req *request.Request
	switch method {
	case "PUT":
		req, _ = cli.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(cli.Bucket),
			Key:    aws.String(key),
		})
	default:
		return "", fmt.Errorf(`unsupported method, only "GET", "PUT"`)
	}

	return cli.presign(req, exp)
}

func (cli *AwsClient) presign(req *request.Request, exp time.Duration) (string, error) {
	url, err := req.Presign(exp)
	if err != nil {
		return "", err
	}
	// url = strings.ReplaceAll(url, c.PresignEndpoint, c.Host)
	return url, nil
}

type PresignedParam struct {
	Key            string
	Expires        time.Duration
	WantedFileName *string
}

func (cli *AwsClient) PresignedGetURL(p PresignedParam) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(cli.Bucket),
		Key:    aws.String(p.Key),
	}

	if p.WantedFileName != nil {
		input.ResponseContentDisposition = aws.String(fmt.Sprintf(`filename="%v"`, *p.WantedFileName))
	}
	req, _ := cli.S3.GetObjectRequest(input)
	return cli.presign(req, p.Expires)
}

func (cli *AwsClient) CopyObjectWithChan(from, to string) chan error {
	ch := make(chan error, 1)
	if from == to {
		close(ch)
	} else {
		cli.UploadPool.Submit(func() {
			defer close(ch)
			ch <- cli.CopyObject(from, to)
		})
	}
	return ch
}

func (cli *AwsClient) CopyObject(from, to string) error {
	_, err := cli.S3.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(cli.Bucket),
		CopySource: aws.String(filepath.Join(cli.Bucket, from)),
		Key:        aws.String(to),
	})
	if err != nil {
		return err
	}
	return nil
}

func NoSuchKeyErr(err error) bool {
	if err == nil {
		return false
	}
	sErr, ok := err.(awserr.RequestFailure)
	if !ok {
		return false
	}

	return sErr.Code() == s3.ErrCodeNoSuchKey
}

func (cli *AwsClient) ListObjects(prefix string) (*s3.ListObjectsOutput, error) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(cli.Bucket),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}
	out, err := cli.S3.ListObjects(input)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (cli *AwsClient) DeleteDir(dir string) error {
	var wg sync.WaitGroup
	_, err := cli.GetObjectListWithPrefix(dir, func(output *s3.ListObjectsOutput, b bool) bool {
		keys := make([]string, len(output.Contents))
		for i, o := range output.Contents {
			keys[i] = *o.Key
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cli.RemoveFiles(keys)
			if err != nil {
				log.Printf("delete objects failed: %v\n", err)
				return
			}
		}()
		return true
	})
	wg.Wait()
	if err != nil {
		return err
	}
	return nil
}

type ObjectInfo struct {
	Key  string
	Size int64
}

type ListObjectsCB = func(output *s3.ListObjectsOutput, b bool) bool

func (cli *AwsClient) GetObjectListWithPrefix(prefix string, f ListObjectsCB) ([]ObjectInfo, error) {
	objectInfoList := make([]ObjectInfo, 0)
	err := cli.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(cli.Bucket),
		Prefix: aws.String(prefix),
	}, func(output *s3.ListObjectsOutput, b bool) bool {
		for _, obj := range output.Contents {
			objectInfoList = append(objectInfoList, ObjectInfo{
				Key:  *obj.Key,
				Size: *obj.Size,
			})
		}
		if f != nil {
			return f(output, b)
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	return objectInfoList, nil
}

func (cli *AwsClient) ListBucket() (res *s3.ListBucketsOutput, err error) {
	res, err = cli.S3.ListBuckets(&s3.ListBucketsInput{})
	return
}

func (cli *AwsClient) ListDir() {

}
