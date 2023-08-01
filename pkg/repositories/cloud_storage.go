package repositories

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"golek_posts_service/pkg/contracts"
	"google.golang.org/api/option"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"strings"
)

type GoogleCloudStorageService struct {
	client     *storage.Client
	keyPath    string
	bucketName string
}

func (g *GoogleCloudStorageService) UploadFile(ctx context.Context, file *multipart.FileHeader, name string) (string, error) {

	bucket := g.client.Bucket(g.bucketName)

	//Open File
	openedFile, err := file.Open()
	if err != nil {
		return "", err
	}

	defer func(openedFile multipart.File) {
		err := openedFile.Close()
		if err != nil {

		}
	}(openedFile)

	//Read File
	fileBytes, err := ioutil.ReadAll(openedFile)
	if err != nil {
		return "", err
	}

	//Write bytes to Bucket Object
	w := bucket.Object(name).NewWriter(ctx)
	_, _ = w.Write(fileBytes)

	err = w.Close()
	if err != nil {
		return "", err
	}

	parsed, err := url.Parse(w.Attrs().Name)
	if err != nil {
		return "", err
	}

	return "https://storage.googleapis.com/" + g.bucketName + "/" + parsed.EscapedPath(), nil
}

func (g *GoogleCloudStorageService) DeleteFile(ctx context.Context, fileUrl string) error {

	bucket := g.client.Bucket(g.bucketName)
	object := bucket.Object(g.extractObjectName(fileUrl))
	_ = object.Delete(ctx)

	return nil
}

func (g *GoogleCloudStorageService) Connect() error {

	client, err := storage.NewClient(context.Background(), option.WithCredentialsFile(g.keyPath))
	if err != nil {
		return err
	}

	g.client = client
	return nil
}

func (g *GoogleCloudStorageService) extractObjectName(urls string) string {
	uri, err := url.Parse(urls)
	if err != nil {
		panic(err)
	}

	path := uri.Path
	// remove the prefix "/ayocode1-bucker/"
	path = path[len(fmt.Sprintf("/%s/", g.bucketName)):]

	// replace the space with "_"
	return strings.Replace(path, "%20", " ", -1)
}

func NewGCStorageService(bucketName string, keyPath string) contracts.ICloudStorageRepo {
	return &GoogleCloudStorageService{client: nil, bucketName: bucketName, keyPath: keyPath}
}
