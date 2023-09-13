package hybridresources

import (
	"context"
	"fmt"
	"log"

	"vm/iam"

	azurestack "github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/resources/mgmt/resources"
	"github.com/Azure/go-autorest/autorest"
)

const (
	errorPrefix = "Cannot create resource group, reason: %v"
)

func getResourceGroupsClient(armEndpoint, tenantID, clientID, clientSecret, subscriptionID string) azurestack.GroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}

	groupsClient := azurestack.NewGroupsClientWithBaseURI(armEndpoint, subscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return groupsClient
}

// CreateResourceGroup creates resource group
func CreateResourceGroup(cntx context.Context, resourceGroupName, location, armEndpoint, tenantID, clientID, clientSecret, subscriptionID string) (name *string, err error) {
	groupClient := getResourceGroupsClient(armEndpoint, tenantID, clientID, clientSecret, subscriptionID)
	rg, errRg := groupClient.CreateOrUpdate(cntx, resourceGroupName, azurestack.Group{Location: &location})
	if errRg == nil {
		name = rg.Name
	}
	err = errRg
	return name, err
}

func DeleteResourceGroup(cntx context.Context, resourceGroupName, armEndpoint, tenantID, clientID, clientSecret, subscriptionID string) (resp autorest.Response, err error) {
	groupClient := getResourceGroupsClient(armEndpoint, tenantID, clientID, clientSecret, subscriptionID)
	future, err := groupClient.Delete(cntx, resourceGroupName)
	if err != nil {
		return resp, err
	}
	err = future.WaitForCompletionRef(cntx, groupClient.Client)
	if err != nil {
		return resp, err
	}
	return future.Result(groupClient)
}
