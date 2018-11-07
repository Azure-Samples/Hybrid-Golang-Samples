package hybridresources

import (
	"context"
	"fmt"

	"iam"
	"log"

	azurestack "github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/resources"
	publicazure "github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/Azure/go-autorest/autorest"
)

const (
	errorPrefix = "Cannot create resource group, reason: %v"
)

func getStackGroupsClient(activeDirectoryEndpoint, activeDirectoryResourceID, armEndpoint, tenantID, clientID, clientSecret, subscriptionID string) azurestack.GroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tenantID, clientID, clientSecret, activeDirectoryResourceID)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}

	groupsClient := azurestack.NewGroupsClientWithBaseURI(armEndpoint, subscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return groupsClient
}

func getPublicGroupsClient(activeDirectoryEndpoint, activeDirectoryResourceID, armEndpoint, tenantID, clientID, clientSecret string, subscriptionID string) publicazure.GroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tenantID, clientID, clientSecret, activeDirectoryResourceID)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}

	groupsClient := publicazure.NewGroupsClientWithBaseURI(armEndpoint, subscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return groupsClient
}

// CreateResourceGroup creates resource group
func CreateResourceGroup(cntx context.Context, rgname, environment, location, activeDirectoryEndpoint, activeDirectoryResourceID, armEndpoint, tenantID, clientID, clientSecret, subscriptionID string) (name *string, err error) {
	if environment == "stack" {
		groupClient := getStackGroupsClient(activeDirectoryEndpoint, activeDirectoryResourceID, armEndpoint, tenantID, clientID, clientSecret, subscriptionID)
		rg, errRg := groupClient.CreateOrUpdate(cntx, rgname, azurestack.Group{Location: &location})
		if errRg == nil {
			name = rg.Name
		}
		err = errRg
		return name, err
	}
	groupClient := getPublicGroupsClient(activeDirectoryEndpoint, activeDirectoryResourceID, armEndpoint, tenantID, clientID, clientSecret, subscriptionID)
	rg, errRg := groupClient.CreateOrUpdate(cntx, rgname, publicazure.Group{Location: &location})
	if errRg == nil {
		name = rg.Name
	}
	err = errRg
	return name, err
}
