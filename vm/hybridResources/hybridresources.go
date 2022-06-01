package hybridresources

import (
	"context"
	"fmt"
	"log"

	"vm/iam"

	azurestack "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

const (
	errorPrefix = "Cannot create resource group, reason: %v"
)

func getResourceGroupsClient(tenantID, clientID, clientSecret, subscriptionID string) (*azurestack.ResourceGroupsClient, error) {
	token, err := iam.GetResourceManagementTokenHybrid(tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}

	return azurestack.NewResourceGroupsClient(subscriptionID, token, nil)
}

// CreateResourceGroup creates resource group
func CreateResourceGroup(cntx context.Context, resourceGroupName, location, tenantID, clientID, clientSecret, subscriptionID string) (name *string, err error) {
	groupClient, err := getResourceGroupsClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return name, err
	}
	rg, errRg := groupClient.CreateOrUpdate(cntx, resourceGroupName, azurestack.ResourceGroup{Location: &location}, nil)
	if errRg == nil {
		name = rg.Name
	}
	err = errRg
	return name, err
}

func DeleteResourceGroup(cntx context.Context, resourceGroupName, tenantID, clientID, clientSecret, subscriptionID string) error {
	groupClient, err := getResourceGroupsClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return err
	}
	future, err := groupClient.BeginDelete(cntx, resourceGroupName, nil)
	if err != nil {
		return err
	}
	_, err = future.PollUntilDone(cntx, nil)
	if err != nil {
		return err
	}
	return nil
}
