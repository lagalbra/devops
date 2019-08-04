package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

func uploadImageToAzure(azStorageAcc, azStorageKey, fileName string) {
	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(azStorageAcc, azStorageKey)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	// Create a random string for the quick start container
	containerName := "containerdevops"

	// From the Azure portal, get your storage account blob service URL endpoint.
	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", azStorageAcc, containerName))
	containerURL := azblob.NewContainerURL(*URL, p)

	fmt.Printf("Creating a container named %s\n", containerName)
	ctx := context.Background() // This example uses a never-expiring context
	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)

	if err != nil {
		if err, ok := err.(azblob.StorageError); ok {
			if err.ServiceCode() != "ContainerAlreadyExists" {
				fmt.Println("Unknown Error creating container", err)
				return
			}
		}
	}

	blobURL := containerURL.NewBlockBlobURL(fileName)
	file, err := os.Open(fileName)

	fmt.Printf("Uploading the file with blob name: %s\n", fileName)
	_, err = azblob.UploadFileToBlockBlob(ctx, file, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize: 4 * 1024 * 1024,
		BlobHTTPHeaders: azblob.BlobHTTPHeaders{
			ContentType: "image/png",
		},
		Parallelism: 16})

	if err != nil {
		fmt.Println("Error uploading!!!", err)
		return
	}
}
