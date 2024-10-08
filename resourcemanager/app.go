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

func printResourceGroups(rgClient *armresources.ResourceGroupsClient) {
	pager := rgClient.NewListPager(nil)
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			fmt.Printf("\nErr can't get next page in resource group list")
			os.Exit(1)
		}
		if resp.ResourceGroupListResult.Value != nil {
			for _, rg := range resp.ResourceGroupListResult.Value {
				fmt.Print(*rg.Name + ", ")
			}
		}
	}
	fmt.Println()
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
	environment, _ := azure.EnvironmentFromURL(config.ResourceManagerEndpointUrl)
	splitEndpoint := strings.Split(environment.ActiveDirectoryEndpoint, "/")
	splitEndpointlastIndex := len(splitEndpoint) - 1
	if splitEndpoint[splitEndpointlastIndex] == "adfs" || splitEndpoint[splitEndpointlastIndex] == "adfs/" {
		config.TenantId = "adfs"
		*disableInstanceDiscovery = true
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
		fmt.Printf("Errr getting token: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Creating resource group")

	var resourceGroupName = "TestGoSampleResourceGroup"

	rgoptions := arm.ClientOptions{ClientOptions: clientOptions}
	rgClient, err := armresources.NewResourceGroupsClient(config.SubscriptionId, cred, &rgoptions)

	if err != nil {
		fmt.Printf("Errr creating resource group client: %s\n", err)
		os.Exit(1)
	}

	param := armresources.ResourceGroup{
		Location: to.Ptr(config.Location),
	}

	_, err = rgClient.CreateOrUpdate(cntx, resourceGroupName, param, nil)

	if err != nil {
		fmt.Printf("\nErrr creating resource group: %s", err)
		os.Exit(1)
	}

	_, err = rgClient.Get(cntx, resourceGroupName, nil)
	if err != nil {
		fmt.Printf("\nErrr no resource group found: %s", err)
		os.Exit(1)
	}

	//print RGs
	// List all the resource groups of an Azure subscription.
	fmt.Println("Listing Resource Groups")
	printResourceGroups(rgClient)

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
		fmt.Println("Listing Resource Groups")
		printResourceGroups(rgClient)
	}
}
