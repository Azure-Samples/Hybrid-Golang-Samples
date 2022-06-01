package hybridresources

import (
	"context"
	"fmt"
	"log"

	"manageddisk/iam"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

const (
	errorPrefix = "Cannot create resource group, reason: %v"
)

func getResourceGroupsClient(tenantID, clientID, certPass, certPath, subscriptionID string) (*armresources.ResourceGroupsClient, error) {
	token, err := iam.GetResourceManagementToken(tenantID, clientID, certPass, certPath)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}

	return armresources.NewResourceGroupsClient(subscriptionID, token, nil)
}

func CreateResourceGroup(cntx context.Context, resourceGroupName, location, certPath, tenantID, clientID, certPass, subscriptionID string) (name *string, err error) {
	groupClient, err := getResourceGroupsClient(tenantID, clientID, certPass, certPath, subscriptionID)
	if err != nil {
		return nil, err
	}
	rg, errRg := groupClient.CreateOrUpdate(cntx, resourceGroupName, armresources.ResourceGroup{Location: &location}, nil)
	if errRg == nil {
		name = rg.Name
	}
	err = errRg
	return name, err
}

func DeleteResourceGroup(cntx context.Context, resourceGroupName, certPath, tenantID, clientID, certPass, subscriptionID string) error {
	groupClient, err := getResourceGroupsClient(certPath, tenantID, clientID, certPass, subscriptionID)
	if err != nil {
		return err
	}

	future, err := groupClient.BeginDelete(cntx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = future.PollUntilDone(cntx, nil)
	return err
}
