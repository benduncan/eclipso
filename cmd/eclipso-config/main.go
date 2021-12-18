package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/benduncan/eclipso/pkg/config"
	"github.com/pelletier/go-toml/v2"
)

func main() {

	file, err := os.ReadFile("domains/phasegrid.net.toml")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	myconfig := config.Config{}
	toml.Unmarshal(file, &myconfig)
	config.ApplyDefaults(&myconfig)

	fmt.Println(myconfig)

	reads3()

}

func reads3() {

	bucket := os.Getenv("S3_BUCKET")

	path := strings.Split(bucket, "s3://")

	fmt.Println("SplitN => ", path[1])

	config.ReadZoneFiles(bucket)

	/*
		sess := session.Must(session.NewSession())

		// Create S3 service client
		svc := s3.New(sess)

		bucket := os.Getenv("S3_BUCKET")

		if bucket == "" {
			log.Fatal("S3_BUCKET env field required")

		}

		// Get the list of items
		resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
		if err != nil {
			log.Fatalf("Unable to list items in bucket %q, %v", bucket, err)
		}

		for _, item := range resp.Contents {

			if strings.HasSuffix(*item.Key, ".toml") {
				fmt.Println("Name:         ", *item.Key)
				fmt.Println("Last modified:", *item.LastModified)
				fmt.Println("Size:         ", *item.Size)
				fmt.Println("Storage class:", *item.StorageClass)
				fmt.Println("")

				buff := &aws.WriteAtBuffer{}

				downloader := s3manager.NewDownloader(sess)

				numBytes, err := downloader.Download(buff,
					&s3.GetObjectInput{
						Bucket: aws.String(bucket),
						Key:    aws.String(*item.Key),
					})

				fmt.Println(numBytes)

				fmt.Println(string(buff.Bytes()))

				myconfig := config.Config{}
				toml.Unmarshal(buff.Bytes(), &myconfig)
				config.ApplyDefaults(&myconfig)

				fmt.Println(myconfig)

				if err != nil {
					log.Fatalf("Unable to download item %q, %v", item, err)
				}

			}

		}

		fmt.Println("Found", len(resp.Contents), "items in bucket", bucket)
		fmt.Println("")

	*/

}
