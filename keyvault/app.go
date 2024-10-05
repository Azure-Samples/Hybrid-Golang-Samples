package main

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profile/p20200901/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/profile/p20200901/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/Azure/go-autorest/autorest/azure"
)

type AzureSpConfig struct {
	ClientId                   string
	CertPass                   string
	CertPath                   string
	ClientSecret               string
	ObjectId                   string
	SubscriptionId             string
	TenantId                   string
	ResourceManagerEndpointUrl string
	Location                   string
}

func main() {
	// Read configuration file for Azure Stack environment details.
	var certConfigFile = "azureCertSpConfig.json"
	var certConfigFilePath = "../" + certConfigFile
	var secretConfigFile = "azureSecretSpConfig.json"
	var secretConfigFilePath = "../" + secretConfigFile
	var config AzureSpConfig
	var data, certData []byte
	var err error
	var certs []*x509.Certificate
	var privateKey crypto.PrivateKey

	//parse flags
	usingSecret := flag.Bool("secret", false, "use secret config file")
	clean := flag.Bool("clean", false, "clean resource groups")
	disableInstanceDiscovery := flag.Bool("disableID", false, "disables instance discovery")
	flag.Parse()

	if *usingSecret {
		goto USINGSECRET
	}

	_, err = os.Stat(certConfigFilePath)
	if err != nil {
		goto USINGSECRET
	}
	data, err = os.ReadFile(certConfigFilePath)
	if err != nil {
		goto USINGSECRET
	}
	err = json.Unmarshal([]byte(data), &config)
	if err != nil {
		goto USINGSECRET
	}

	certData, _ = os.ReadFile(config.CertPath)
	certs, privateKey, err = azidentity.ParseCertificates(certData, []byte(config.CertPass))
	if err != nil {
		fmt.Println("Unable to parse Certificate")
		goto USINGSECRET
	}

	goto USINGCERT

USINGSECRET:
	_, err = os.Stat(secretConfigFilePath)
	if err != nil {
		fmt.Printf("The configuration files, %s & %s, don't exist.", secretConfigFilePath, certConfigFilePath)
		os.Exit(1)
	}
	data, err = os.ReadFile(secretConfigFilePath)
	if err != nil {
		fmt.Printf("Failed to read configuration file %s & %s.", secretConfigFile, certConfigFilePath)
		os.Exit(1)
	}

	err = json.Unmarshal([]byte(data), &config)
	if err != nil {
		fmt.Printf("Failed to unmarshal data from %s & %s.", secretConfigFile, certConfigFilePath)
		os.Exit(1)
	}

USINGCERT:
	cntx := context.Background()
	environment, err := azure.EnvironmentFromURL(config.ResourceManagerEndpointUrl)
	if err != nil {
		fmt.Printf("Failed to get environment from url: %s", err)
		os.Exit(1)
	}
	splitEndpoint := strings.Split(environment.ActiveDirectoryEndpoint, "/")
	splitEndpointlastIndex := len(splitEndpoint) - 1
	adminTenantId := config.TenantId
	if splitEndpoint[splitEndpointlastIndex] == "adfs" || splitEndpoint[splitEndpointlastIndex] == "adfs/" {
		*disableInstanceDiscovery = true
		config.TenantId = "adfs"
	}

	fmt.Println("Creating credential and getting token")

	cloudConfig := cloud.Configuration{ActiveDirectoryAuthorityHost: environment.ActiveDirectoryEndpoint, Services: map[cloud.ServiceName]cloud.ServiceConfiguration{cloud.ResourceManager: {Endpoint: environment.ResourceManagerEndpoint, Audience: environment.TokenAudience}}}

	clientOptions := policy.ClientOptions{Cloud: cloudConfig}

	var cred azcore.TokenCredential
	if *usingSecret {
		options := azidentity.ClientSecretCredentialOptions{ClientOptions: clientOptions, DisableInstanceDiscovery: *disableInstanceDiscovery}
		cred, err = azidentity.NewClientSecretCredential(config.TenantId, config.ClientId, config.ClientSecret, &options)
		if err != nil {
			fmt.Printf("Error getting client secret cred: %s\n", err)
			os.Exit(1)
		}
		_, err = cred.GetToken(cntx, policy.TokenRequestOptions{Scopes: []string{environment.TokenAudience + "/.default"}})
	} else {
		options := azidentity.ClientCertificateCredentialOptions{ClientOptions: clientOptions, DisableInstanceDiscovery: *disableInstanceDiscovery}
		cred, err = azidentity.NewClientCertificateCredential(config.TenantId, config.ClientId, certs, privateKey, &options)
		if err != nil {
			fmt.Printf("Error getting client certificate cred: %s\n", err)
			os.Exit(1)
		}
		_, err = cred.GetToken(cntx, policy.TokenRequestOptions{Scopes: []string{environment.TokenAudience + "/.default"}})
	}

	if err != nil {
		fmt.Printf("Error getting token: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Creating resource group")

	var resourceGroupName = "TestGoKVSampleResourceGroup"

	rgoptions := arm.ClientOptions{ClientOptions: clientOptions}
	rgClient, err := armresources.NewResourceGroupsClient(config.SubscriptionId, cred, &rgoptions)

	if err != nil {
		fmt.Printf("Error creating resource group client: %s\n", err)
		os.Exit(1)
	}

	param := armresources.ResourceGroup{
		Location: to.Ptr(config.Location),
	}

	_, err = rgClient.CreateOrUpdate(cntx, resourceGroupName, param, nil)
	if err != nil {
		fmt.Printf("\nError creating resource group: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Creating Key Vault client")
	kvClient, err := armkeyvault.NewVaultsClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("\nError creating KV client: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Printing Key Vaults")
	pager := kvClient.NewListPager(armkeyvault.Enum10ResourceTypeEqMicrosoftKeyVaultVaults, armkeyvault.Enum11TwoThousandFifteen1101, nil)
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			fmt.Printf("\nErr can't get next page in KV list\n")
			os.Exit(1)
		}
		if resp.ResourceListResult.Value != nil {
			for _, kv := range resp.ResourceListResult.Value {
				fmt.Print(*kv.Name + ", ")
			}
		}
	}
	fmt.Println()

	var kvName = "gotestkeyvault"
	//Check name currently not supported
	// fmt.Println("Checking name availability")

	// availability, err := kvClient.CheckNameAvailability(context.Background(), armkeyvault.VaultCheckNameAvailabilityParameters{Name: &kvName}, nil)
	// if err != nil {
	// 	fmt.Printf("\nErr checking KV name availability: %s", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("The account %s is available: %t\n", kvName, *availability.NameAvailable)
	// if !*availability.NameAvailable {
	// 	fmt.Printf("Detailed message: %s\n", *availability.Message)
	// 	os.Exit(1)
	// }

	var skuFamily = armkeyvault.SKUFamilyA
	var skuname = armkeyvault.SKUNameStandard
	cntxTimeout1, cancel := context.WithTimeout(cntx, 30*time.Second)
	defer cancel()
	fmt.Println("Creating Key Vault")
	result, err := kvClient.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		kvName,
		armkeyvault.VaultCreateOrUpdateParameters{
			Location: to.Ptr(config.Location),
			Properties: &armkeyvault.VaultProperties{
				TenantID: &adminTenantId,
				SKU: &armkeyvault.SKU{
					Family: &skuFamily,
					Name:   &skuname,
				},
				AccessPolicies: []*armkeyvault.AccessPolicyEntry{{
					ObjectID: to.Ptr(config.ObjectId),
					TenantID: &adminTenantId,
					Permissions: &armkeyvault.Permissions{
						Secrets:      []*armkeyvault.SecretPermissions{to.Ptr(armkeyvault.SecretPermissionsAll)},
						Keys:         []*armkeyvault.KeyPermissions{to.Ptr(armkeyvault.KeyPermissionsAll)},
						Storage:      []*armkeyvault.StoragePermissions{to.Ptr(armkeyvault.StoragePermissionsAll)},
						Certificates: []*armkeyvault.CertificatePermissions{to.Ptr(armkeyvault.CertificatePermissionsAll)},
					},
				}},
			},
		},
		nil,
	)
	if err != nil {
		fmt.Printf("\nErr creating KV: %s\n", err)
		os.Exit(1)
	}
	result.PollUntilDone(cntxTimeout1, nil)

	fmt.Println("Printing Key Vaults")
	pager1 := kvClient.NewListPager(armkeyvault.Enum10ResourceTypeEqMicrosoftKeyVaultVaults, armkeyvault.Enum11TwoThousandFifteen1101, nil)
	for pager1.More() {
		resp, err := pager1.NextPage(context.Background())
		if err != nil {
			fmt.Printf("\nErr can't get next page in KV list\n")
			os.Exit(1)
		}
		if resp.ResourceListResult.Value != nil {
			for _, kv := range resp.ResourceListResult.Value {
				fmt.Print(*kv.Name + ", ")
			}
		}
	}
	fmt.Println()

	fmt.Println("Creating Secret Client")
	secClient, err := armkeyvault.NewSecretsClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("\nErr creating secrets client: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Creating secret in Key Vault")
	var secretName = "testgokey"
	var secretValue = "testvalue"
	_, err = secClient.CreateOrUpdate(
		context.Background(),
		resourceGroupName,
		kvName,
		secretName,
		armkeyvault.SecretCreateOrUpdateParameters{
			Properties: &armkeyvault.SecretProperties{
				Value: &secretValue,
			},
		},
		nil,
	)
	if err != nil {
		fmt.Printf("\nErr creating secret: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Getting secret from Key Vault")
	secresp, err := secClient.Get(context.Background(), resourceGroupName, kvName, secretName, nil)

	if err != nil {
		fmt.Printf("\nErr getting secret %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Secret retrieved. Name: %s\n", *secresp.Name)

	fmt.Println("Deleting Key Vault")
	cntxTimeout, cancel := context.WithTimeout(cntx, 120*time.Second)
	defer cancel()
	kvClient.Delete(cntxTimeout, resourceGroupName, kvName, nil)
	if err != nil {
		fmt.Printf("Failed to delete keyvault: %s\n", resourceGroupName)
		os.Exit(1)
	}

	if *clean {
		fmt.Println("Deleting resource group")
		result, err := rgClient.BeginDelete(context.Background(), resourceGroupName, nil)
		if err != nil {
			fmt.Printf("Failed to delete resource group: %s\n", resourceGroupName)
			os.Exit(1)
		}

		cntxTimeout, cancel := context.WithTimeout(cntx, 300*time.Second)
		defer cancel()
		_, err = result.PollUntilDone(cntxTimeout, nil)
		if err != nil {
			fmt.Println("Timed out when deleting resource group")
			os.Exit(1)
		}
	}
}
