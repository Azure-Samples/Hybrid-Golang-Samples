package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	hybridkeyvault "keyvault/hybridkeyvault"
	hybridresources "keyvault/hybridresources"

	"github.com/Azure/go-autorest/autorest/azure"
)

var (
	resourceGroupName      = "azure-sample-golang-keyvault"
	keyVaultName           = "azure-sample-keyvault"
	keyVaultWithPolicyName = "azure-sample-policyKV"
)

type AzureAppSpConfig struct {
	ClientId           string
	ClientSecret       string
	ClientObjectId     string
	SubscriptionId     string
	TenantId           string
	ResourceManagerUrl string
	Location           string
}

func main() {
	// Read configuration file for Azure Stack environment details.
	var configFile = "azureAppSpConfig.json"
	var configFilePath = "../" + configFile
	_, err := os.Stat(configFilePath)
	if err != nil {
		log.Fatalf("The configuration file, %s, doesn't exist.", configFilePath)
	}
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Failed to read configuration file %s.", configFile)
	}
	var config AzureAppSpConfig
	err = json.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal data from %s.", configFile)
	}

	cntx := context.Background()

	// Determine whether the environment is ADFS or AAD.
	environment, _ := azure.EnvironmentFromURL(config.ResourceManagerUrl)
	splitEndpoint := strings.Split(environment.ActiveDirectoryEndpoint, "/")
	splitEndpointlastIndex := len(splitEndpoint) - 1
	tenantUUID := config.TenantId
	if splitEndpoint[splitEndpointlastIndex] == "adfs" || splitEndpoint[splitEndpointlastIndex] == "adfs/" {
		config.TenantId = "adfs"
	}

	if len(os.Args) == 2 && os.Args[1] == "clean" {
		fmt.Printf("Deleting resource group '%s'...\n", resourceGroupName)
		//Create a resource group on Azure Stack
		_, err := hybridresources.DeleteResourceGroup(
			cntx,
			resourceGroupName,
			config.ResourceManagerUrl,
			config.TenantId,
			config.ClientId,
			config.ClientSecret,
			config.SubscriptionId)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			fmt.Printf("Successfully deleted resource group '%s'.\n", resourceGroupName)
		}
		return
	}

	fmt.Printf("Creating resource group '%s'...\n", resourceGroupName)
	// Create a resource group on Azure Stack
	_, errResourceGroup := hybridresources.CreateResourceGroup(
		cntx,
		resourceGroupName,
		config.Location,
		config.ResourceManagerUrl,
		config.TenantId,
		config.ClientId,
		config.ClientSecret,
		config.SubscriptionId)
	if errResourceGroup != nil {
		log.Fatal(errResourceGroup.Error())
	} else {
		fmt.Printf("Successfully created resource group '%s'.\n", resourceGroupName)
	}

	fmt.Printf("Creating key vault '%s'...\n", keyVaultName)
	// Create a key vault.
	keyVaultObject, errCreateKeyVault := hybridkeyvault.CreateVault(
		cntx,
		keyVaultName,
		tenantUUID,
		config.TenantId,
		config.ClientId,
		config.ClientSecret,
		config.ResourceManagerUrl,
		config.SubscriptionId,
		resourceGroupName,
		config.Location,
	)
	if errCreateKeyVault != nil {
		log.Fatal(errCreateKeyVault.Error())
	} else {
		fmt.Printf("Successfully created key vault '%s'.\n", *keyVaultObject.Name)
	}

	fmt.Printf("Getting key vault '%s'...\n", keyVaultName)
	// Get the created key vault.
	GotKeyVault, errGetKeyVault := hybridkeyvault.GetVault(
		cntx,
		keyVaultName,
		config.TenantId,
		config.ClientId,
		config.ClientSecret,
		config.ResourceManagerUrl,
		config.SubscriptionId,
		resourceGroupName,
		config.Location,
	)
	if errGetKeyVault != nil {
		log.Fatal(errGetKeyVault.Error())
	} else {
		fmt.Printf("Successfully got key vault '%s'.\n", *GotKeyVault.Name)
	}

	fmt.Printf("Creating key vault '%s' with policies...\n", keyVaultWithPolicyName)
	// Create a new key vault with policies.
	keyVaultWithPoliciesObject, errCreateKeyVaultWithPolicies := hybridkeyvault.CreateVaultWithPolicies(
		cntx,
		keyVaultWithPolicyName,
		tenantUUID,
		config.TenantId,
		config.ClientObjectId,
		config.ClientId,
		config.ClientSecret,
		config.ResourceManagerUrl,
		config.SubscriptionId,
		resourceGroupName,
		config.Location,
	)
	if errCreateKeyVaultWithPolicies != nil {
		log.Fatal(errCreateKeyVaultWithPolicies.Error())
	} else {
		fmt.Printf("Successfully created key vault '%s' with policies.\n", *keyVaultWithPoliciesObject.Name)
	}

	fmt.Printf("Setting new permissions for key vault '%s'...\n", keyVaultName)
	// Set new key vault permissions.
	KeyVaultWithSetPermissions, errSetKeyVault := hybridkeyvault.SetVaultPermissions(
		cntx,
		keyVaultName,
		tenantUUID,
		config.TenantId,
		config.ClientObjectId,
		config.ClientId,
		config.ClientSecret,
		config.ResourceManagerUrl,
		config.SubscriptionId,
		resourceGroupName,
		config.Location,
	)
	if errSetKeyVault != nil {
		log.Fatal(errSetKeyVault.Error())
	} else {
		fmt.Printf("Successfully set new permissions for key vault '%s'.\n", *KeyVaultWithSetPermissions.Name)
	}

	fmt.Printf("Setting deployment permissions for key vault '%s'...\n", keyVaultWithPolicyName)
	// Set keyvault deployment permission.
	keyVaultWithDeploymentPermission, errSetKeyVaultWithDeploymentPermission := hybridkeyvault.SetVaultDeploymentPermission(
		cntx,
		keyVaultWithPolicyName,
		tenantUUID,
		config.TenantId,
		config.ClientObjectId,
		config.ClientId,
		config.ClientSecret,
		config.ResourceManagerUrl,
		config.SubscriptionId,
		resourceGroupName,
		config.Location,
	)
	if errSetKeyVaultWithDeploymentPermission != nil {
		log.Fatal(errSetKeyVaultWithDeploymentPermission.Error())
	} else {
		fmt.Printf("Successfully set deployment permissions for key vault '%s'.\n", *keyVaultWithDeploymentPermission.Name)
	}

	fmt.Printf("Getting list of key vaults...\n")
	// Get list of key vaults for configured subscription.
	hybridkeyvault.GetVaults(
		config.TenantId,
		config.ClientId,
		config.ClientSecret,
		config.ResourceManagerUrl,
		config.SubscriptionId,
		resourceGroupName,
	)

	fmt.Printf("Deleting key vault '%s'...\n", keyVaultName)
	// Set keyvault deployment permission.
	_, errDeleteKeyVault := hybridkeyvault.DeleteVault(
		cntx,
		keyVaultName,
		config.TenantId,
		config.ClientId,
		config.ClientSecret,
		config.ResourceManagerUrl,
		config.SubscriptionId,
		resourceGroupName,
	)
	if errDeleteKeyVault != nil {
		log.Fatal(errDeleteKeyVault.Error())
	} else {
		fmt.Printf("Succesfully deleted key vault '%s'.\n", keyVaultName)
	}

	return
}
