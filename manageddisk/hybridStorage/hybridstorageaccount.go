package hybridstorage

import (
	"context"
	"fmt"
	"log"

	"manageddisk/iam"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

const (
	errorPrefix = "Cannot create storage account, reason: %v"
)

func getStorageAccountsClient(certPath, tenantID, clientID, certPass, subscriptionID string) (*armstorage.AccountsClient, error) {
	token, err := iam.GetResourceManagementToken(tenantID, clientID, certPass, certPath)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}

	return armstorage.NewAccountsClient(subscriptionID, token, nil)
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount(cntx context.Context, accountName, rgName, location, certPath, tenantID, clientID, certPass, subscriptionID string) (s armstorage.Account, err error) {
	storageAccountsClient, err := getStorageAccountsClient(certPath, tenantID, clientID, certPass, subscriptionID)
	if err != nil {
		return s, err
	}
	result, err := storageAccountsClient.CheckNameAvailability(
		cntx,
		armstorage.AccountCheckNameAvailabilityParameters{
			Name: to.Ptr(accountName),
			Type: to.Ptr("Microsoft.Storage/storageAccounts"),
		},
		nil)
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
		nil)
	if err != nil {
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}

	resp, err := future.PollUntilDone(cntx, nil)
	if err != nil {
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, fmt.Sprintf("cannot get the storage account create future response: %v", err)))
	}

	return resp.Account, nil
}
