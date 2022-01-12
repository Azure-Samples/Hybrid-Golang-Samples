package hybridkeyvault

import (
	"context"
	"fmt"
	"log"

	"keyvault/iam"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/keyvault/mgmt/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	uuid "github.com/gofrs/uuid"
)

const (
	errorPrefix = "Cannot create resource group, reason: %v"
)

// Get the keyvault.VaultsClient
func getVaultsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID string) keyvault.VaultsClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}

	keyvaultClient := keyvault.NewVaultsClientWithBaseURI(armEndpoint, subscriptionID)
	keyvaultClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return keyvaultClient
}

// CreateVault creates a new vault
func CreateVault(ctx context.Context, vaultName, tenantUUID, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, resourceGroup, location string) (vault keyvault.Vault, err error) {
	vaultsClient := getVaultsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)

	tenantUUIDObject, err := uuid.FromString(tenantUUID)
	if err != nil {
		return vault, err
	}

	future, err := vaultsClient.CreateOrUpdate(
		ctx,
		resourceGroup,
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &keyvault.VaultProperties{
				TenantID: &tenantUUIDObject,
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{},
			},
		},
	)

	if err != nil {
		return vault, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletionRef(ctx, vaultsClient.Client)
	if err != nil {
		return vault, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	return future.Result(vaultsClient)
}

// GetVault returns an existing vault
func GetVault(ctx context.Context, vaultName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, resourceGroup, location string) (keyvault.Vault, error) {
	vaultsClient := getVaultsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	return vaultsClient.Get(ctx, resourceGroup, vaultName)
}

// CreateVaultWithPolicies creates a new Vault with policies granting access to the specified service principal.
func CreateVaultWithPolicies(ctx context.Context, vaultName, tenantUUID, tenantID, clientObjectID, clientID, clientSecret, armEndpoint, subscriptionID, resourceGroup, location string) (vault keyvault.Vault, err error) {
	vaultsClient := getVaultsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)

	tenantUUIDObject, err := uuid.FromString(tenantUUID)
	if err != nil {
		return
	}

	var accessPolicyList []keyvault.AccessPolicyEntry
	accessPolicy := keyvault.AccessPolicyEntry{
		TenantID: &tenantUUIDObject,
		Permissions: &keyvault.Permissions{
			Keys: &[]keyvault.KeyPermissions{
				keyvault.KeyPermissionsCreate,
			},
			Secrets: &[]keyvault.SecretPermissions{
				keyvault.SecretPermissionsSet,
			},
		},
	}
	if clientObjectID != "" {
		accessPolicy.ObjectID = to.StringPtr(clientObjectID)
		accessPolicyList = append(accessPolicyList, accessPolicy)
	}

	future, err := vaultsClient.CreateOrUpdate(
		ctx,
		resourceGroup,
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &keyvault.VaultProperties{
				AccessPolicies:           &accessPolicyList,
				EnabledForDiskEncryption: to.BoolPtr(true),
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				TenantID: &tenantUUIDObject,
			},
		})

	if err != nil {
		return vault, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletionRef(ctx, vaultsClient.Client)
	if err != nil {
		return vault, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	return future.Result(vaultsClient)
}

// SetVaultPermissions adds an access policy permitting this app's Client ID to manage keys and secrets.
func SetVaultPermissions(ctx context.Context, vaultName, tenantUUID, tenantID, clientObjectID, clientID, clientSecret, armEndpoint, subscriptionID, resourceGroup, location string) (vault keyvault.Vault, err error) {
	vaultsClient := getVaultsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)

	tenantUUIDObject, err := uuid.FromString(tenantUUID)
	if err != nil {
		return vault, err
	}

	future, err := vaultsClient.CreateOrUpdate(
		ctx,
		resourceGroup,
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &keyvault.VaultProperties{
				TenantID: &tenantUUIDObject,
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{
					{
						ObjectID: &clientObjectID,
						TenantID: &tenantUUIDObject,
						Permissions: &keyvault.Permissions{
							Keys: &[]keyvault.KeyPermissions{
								keyvault.KeyPermissionsGet,
								keyvault.KeyPermissionsList,
								keyvault.KeyPermissionsCreate,
							},
							Secrets: &[]keyvault.SecretPermissions{
								keyvault.SecretPermissionsGet,
								keyvault.SecretPermissionsList,
							},
						},
					},
				},
			},
		},
	)

	if err != nil {
		return vault, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletionRef(ctx, vaultsClient.Client)
	if err != nil {
		return vault, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	return future.Result(vaultsClient)
}

// Updates a key vault to enable deployments and add permissions to the application
func SetVaultDeploymentPermission(ctx context.Context, vaultName, tenantUUID, tenantID, clientObjectID, clientID, clientSecret, armEndpoint, subscriptionID, resourceGroup, location string) (vault keyvault.Vault, err error) {
	vaultsClient := getVaultsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	tenantUUIDObject, err := uuid.FromString(tenantUUID)
	if err != nil {
		return vault, err
	}

	future, err := vaultsClient.CreateOrUpdate(
		ctx,
		resourceGroup,
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &keyvault.VaultProperties{
				TenantID:                     &tenantUUIDObject,
				EnabledForDeployment:         to.BoolPtr(true),
				EnabledForTemplateDeployment: to.BoolPtr(true),
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{
					{
						ObjectID: &clientObjectID,
						TenantID: &tenantUUIDObject,
						Permissions: &keyvault.Permissions{
							Keys: &[]keyvault.KeyPermissions{
								keyvault.KeyPermissionsGet,
								keyvault.KeyPermissionsList,
								keyvault.KeyPermissionsCreate,
							},
							Secrets: &[]keyvault.SecretPermissions{
								keyvault.SecretPermissionsGet,
								keyvault.SecretPermissionsSet,
								keyvault.SecretPermissionsList,
							},
						},
					},
				},
			},
		},
	)

	if err != nil {
		return vault, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletionRef(ctx, vaultsClient.Client)
	if err != nil {
		return vault, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	return future.Result(vaultsClient)
}

// GetVaults lists all key vaults in a subscription
func GetVaults(tenantID, clientID, clientSecret, armEndpoint, subscriptionID, resourceGroup string) {
	vaultsClient := getVaultsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)

	fmt.Println("Getting all vaults in subscription")
	for subList, err := vaultsClient.ListComplete(context.Background(), nil); subList.NotDone(); err = subList.Next() {
		if err != nil {
			log.Printf("failed to get list of vaults: %v", err)
		}
		fmt.Printf("\t%s\n", *subList.Value().Name)
	}

	fmt.Println("Getting all vaults in resource group")
	for rgList, err := vaultsClient.ListByResourceGroupComplete(context.Background(), resourceGroup, nil); rgList.NotDone(); err = rgList.Next() {
		if err != nil {
			log.Printf("failed to get list of vaults: %v", err)
		}
		fmt.Printf("\t%s\n", *rgList.Value().Name)
	}
}

// DeleteVault deletes an existing vault
func DeleteVault(ctx context.Context, vaultName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, resourceGroup string) (response autorest.Response, err error) {
	vaultsClient := getVaultsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	return vaultsClient.Delete(ctx, resourceGroup, vaultName)
}
