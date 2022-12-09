package ceph

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"gos3/config"
	"testing"
)

func TestMain(m *testing.M) {
	err := InitAwsClient(config.OSSConfig{
		Endpoint:                  "10.198.30.156",
		AccessKeyID:               "spe",
		AccessKeySecret:           "Spe2077@#$%",
		Bucket:                    "spe_data",
		S3ForcePathStyle:          true,
		InsecureSkipVerify:        true,
		DisableEndpointHostPrefix: true,
	})
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestAwsClient_ListBucket(t *testing.T) {
	res, err := Client.ListBucket()
	if err != nil {
		t.Log(err)
	}
	for _, v := range res.Buckets {
		fmt.Print(v)
	}
}

func TestAwsClient_ListDir(t *testing.T) {
	res, err := Client.S3.ListObjects(&s3.ListObjectsInput{
		Bucket:    aws.String(Client.Bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String("test/tasks/"),
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}
