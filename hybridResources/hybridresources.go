package hybridresources

import (
	"context"
	"fmt"
	"log"

	"Hybrid-Compute-Go-Create-VM/iam"

	azurestack "github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/resources/mgmt/resources"
	"github.com/Azure/go-autorest/autorest"
)

const (
	errorPrefix = "Cannot create resource group, reason: %v"
)

func getStackGroupsClient(armEndpoint, tenantID, clientID, clientSecret, subscriptionID string) azurestack.GroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}

	groupsClient := azurestack.NewGroupsClientWithBaseURI(armEndpoint, subscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return groupsClient
}

// CreateResourceGroup creates resource group
func CreateResourceGroup(cntx context.Context, rgname, location, armEndpoint, tenantID, clientID, clientSecret, subscriptionID string) (name *string, err error) {
	groupClient := getStackGroupsClient(armEndpoint, tenantID, clientID, clientSecret, subscriptionID)
	rg, errRg := groupClient.CreateOrUpdate(cntx, rgname, azurestack.Group{Location: &location})
	if errRg == nil {
		name = rg.Name
	}
	err = errRg
	return name, err
}
