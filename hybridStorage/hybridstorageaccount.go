package hybridStorage

import (
	"context"
	"fmt"

	"log"

	"../iam"

	"github.com/Azure/azure-sdk-for-go/profiles/2018-03-01/storage/mgmt/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	errorPrefix = "Cannot create storage account, reason: %v"
)

func getStorageAccountsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID string) storage.AccountsClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	storageAccountsClient := storage.NewAccountsClientWithBaseURI(armEndpoint, subscriptionID)
	storageAccountsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return storageAccountsClient
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount(cntx context.Context, accountName, rgName, location, tenantID, clientID, clientSecret, armEndpoint, subscriptionID string) (s storage.Account, err error) {
	storageAccountsClient := getStorageAccountsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
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
			Location: to.StringPtr(location),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
		})
	if err != nil {
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletion(cntx, storageAccountsClient.Client)
	if err != nil {
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, fmt.Sprintf("cannot get the storage account create future response: %v", err)))
	}
	return future.Result(storageAccountsClient)
}
