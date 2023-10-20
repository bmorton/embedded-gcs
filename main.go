package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"google.golang.org/api/option"
)

func main() {
	opts := fakestorage.Options{
		Scheme:     "http",
		Port:       8080,
		PublicHost: "0.0.0.0:8080",
		InitialObjects: []fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName: "test-bucket",
					Name:       "test-object",
				},
				Content: []byte("test-content"),
			},
		},
	}
	server, err := fakestorage.NewServerWithOptions(opts)
	if err != nil {
		panic(err)
	}

	interruptCh := make(chan interface{}, 1)

	err = os.Setenv("STORAGE_EMULATOR_HOST", "http://0.0.0.0:8080/storage/v1/")
	if err != nil {
		panic(err)
	}
	client, err := storage.NewClient(context.TODO(), option.WithEndpoint("http://0.0.0.0:8080/storage/v1/"))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	const (
		bucketName = "test-bucket"
		fileKey    = "test-object"
	)
	data, err := downloadFile(client, bucketName, fileKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("contents of %s/%s: %s\n", bucketName, fileKey, data)

	<-interruptCh
	server.Stop()
}

func downloadFile(client *storage.Client, bucketName, fileKey string) ([]byte, error) {
	reader, err := client.Bucket(bucketName).Object(fileKey).NewReader(context.TODO())
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
