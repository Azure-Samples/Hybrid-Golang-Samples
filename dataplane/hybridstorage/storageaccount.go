package storage

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"dataplane/iam"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/storage/mgmt/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	errorPrefix = "Cannot create storage account, reason: %v"
)

func getStorageAccountKey(cntx context.Context, storageAccountsClient storage.AccountsClient, resourceGroupName, storageAccountName string) (key string, err error) {
	listKeys, err := storageAccountsClient.ListKeys(
		cntx,
		resourceGroupName,
		storageAccountName)
	if err != nil {
		return key, fmt.Errorf("cannot list storage account keys: %v", err)
	}
	storageAccountKeys := *listKeys.Keys
	key = *storageAccountKeys[0].Value
	return key, err
}

// UploadDataToContainer uploads data to a container
func UploadDataToContainer(cntx context.Context, containerURL azblob.ContainerURL, blobFileAddress string) (err error) {
	_, err = containerURL.Create(cntx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		return fmt.Errorf("cannot create container: %v", err)
	}
	blobFileAddress = filepath.FromSlash(blobFileAddress)
	blobFileAddressSplit := strings.Split(blobFileAddress, string(os.PathSeparator))
	blobFileName := blobFileAddressSplit[len(blobFileAddressSplit)-1]
	blobURL := containerURL.NewBlockBlobURL(blobFileName)
	file, err := os.Open(blobFileAddress)
	if err != nil {
		return fmt.Errorf("cannot read blob file: %v", err)
	}
	_, err = azblob.UploadFileToBlockBlob(cntx, file, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16})
	return err
}

// GetDataplaneURL returns dataplane URL
func GetDataplaneURL(cntx context.Context, storageAccountsClient storage.AccountsClient, storageEndpointSuffix, storageAccountName, resourceGroupName, storageContainerName string) (containerURL azblob.ContainerURL, err error) {
	storageAccountKey, err := getStorageAccountKey(cntx, storageAccountsClient, resourceGroupName, storageAccountName)
	if err != nil {
		return containerURL, fmt.Errorf("cannot get stroage account key: %v", err)
	}
	credential, err := azblob.NewSharedKeyCredential(storageAccountName, storageAccountKey)
	if err != nil {
		return containerURL, fmt.Errorf("cannot create credential for storage account: %v", err)
	}
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	URL, err := url.Parse(fmt.Sprintf("https://%s.blob.%s/%s", storageAccountName, storageEndpointSuffix, storageContainerName))
	if err != nil {
		return containerURL, fmt.Errorf("cannot create container URL: %v", err)
	}
	containerURL = azblob.NewContainerURL(*URL, pipeline)
	return containerURL, err
}

// GetStorageAccountsClient creates a new storage account client
func GetStorageAccountsClient(tenantID, clientID, certPass, armEndpoint, certPath, subscriptionID string) storage.AccountsClient {
	token, err := iam.GetResourceManagementToken(tenantID, clientID, certPass, armEndpoint, certPath)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	storageAccountsClient := storage.NewAccountsClientWithBaseURI(armEndpoint, subscriptionID)
	storageAccountsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return storageAccountsClient
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount(cntx context.Context, storageAccountsClient storage.AccountsClient, accountName, rgName, location string) (s storage.Account, err error) {
	result, err := storageAccountsClient.CheckNameAvailability(
		cntx,
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	if err != nil {
		return s, fmt.Errorf(errorPrefix, err)
	}
	if *result.NameAvailable != true {
		return s, fmt.Errorf(errorPrefix, fmt.Sprintf("storage account name [%v] not available", accountName))
	}
	future, err := storageAccountsClient.Create(
		cntx,
		rgName,
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Location:                          to.StringPtr(location),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
		})
	if err != nil {
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletionRef(cntx, storageAccountsClient.Client)
	if err != nil {
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, fmt.Sprintf("cannot get the storage account create future response: %v", err)))
	}
	return future.Result(storageAccountsClient)
}
