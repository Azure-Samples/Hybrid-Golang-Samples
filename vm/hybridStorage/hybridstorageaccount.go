package hybridStorage

import (
	"context"
	"fmt"
	"log"

	"vm/iam"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

const (
	errorPrefix = "Cannot create storage account, reason: %v"
)

func getStorageAccountsClient(tenantID, clientID, clientSecret, subscriptionID string) (*armstorage.AccountsClient, error) {
	token, err := iam.GetResourceManagementTokenHybrid(tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	return armstorage.NewAccountsClient(subscriptionID, token, nil)
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount(cntx context.Context, accountName, rgName, location, tenantID, clientID, clientSecret, subscriptionID string) (s armstorage.Account, err error) {
	storageAccountsClient, err := getStorageAccountsClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return s, err
	}
	result, err := storageAccountsClient.CheckNameAvailability(
		cntx,
		armstorage.AccountCheckNameAvailabilityParameters{
			Name: to.Ptr(accountName),
			Type: to.Ptr("Microsoft.Storage/storageAccounts"),
		},
		nil,
	)
	if err != nil {
		return s, fmt.Errorf(errorPrefix, err)
	}
	if *result.NameAvailable != true {
		return s, fmt.Errorf(errorPrefix, fmt.Sprintf("storage account name [%v] not available", accountName))
	}
	future, err := storageAccountsClient.BeginCreate(
		cntx,
		rgName,
		accountName,
		armstorage.AccountCreateParameters{
			SKU: &armstorage.SKU{
				Name: to.Ptr(armstorage.SKUNameStandardLRS),
			},
			Location:   to.Ptr(location),
			Properties: &armstorage.AccountPropertiesCreateParameters{},
		},
		nil,
	)
	if err != nil {
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	resp, err := future.PollUntilDone(cntx, nil)
	if err != nil {
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, fmt.Sprintf("cannot get the storage account create future response: %v", err)))
	}
	return resp.Account, nil
}
