package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	hybridresources "dataplane/hybridresources"
	hybridstorage "dataplane/hybridstorage"

	"github.com/Azure/go-autorest/autorest/azure"
)

var (
	storageAccountName   = "samplestacc"
	resourceGroupName    = "azure-sample-rg"
	storageContainerName = "samplecontainer"
)

type AzureCertSpConfig struct {
	ClientId           string
	CertPass           string
	CertPath           string
	ClientObjectId     string
	SubscriptionId     string
	TenantId           string
	ResourceManagerUrl string
	Location           string
}

func main() {
	// Read configuration file for Azure Stack environment details.
	var configFile = "azureCertSpConfig.json"
	var configFilePath = "../" + configFile
	_, err := os.Stat(configFilePath)
	if err != nil {
		log.Fatalf("The configuration file, %s, doesn't exist.", configFilePath)
	}
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Failed to read configuration file %s.", configFile)
	}
	var config AzureCertSpConfig
	err = json.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal data from %s.", configFile)
	}

	storageEndpointSuffix := strings.TrimRight(config.ResourceManagerUrl[strings.Index(config.ResourceManagerUrl, ".")+1:], "/")

	cntx := context.Background()
	environment, _ := azure.EnvironmentFromURL(config.ResourceManagerUrl)
	splitEndpoint := strings.Split(environment.ActiveDirectoryEndpoint, "/")
	splitEndpointlastIndex := len(splitEndpoint) - 1
	if splitEndpoint[splitEndpointlastIndex] == "adfs" || splitEndpoint[splitEndpointlastIndex] == "adfs/" {
		config.TenantId = "adfs"
	}

	if len(os.Args) == 2 && os.Args[1] == "clean" {
		fmt.Printf("Deleting resource group '%s'...\n", resourceGroupName)
		//Create a resource group on Azure Stack
		_, err := hybridresources.DeleteResourceGroup(
			cntx,
			resourceGroupName,
			config.CertPath,
			config.ResourceManagerUrl,
			config.TenantId,
			config.ClientId,
			config.CertPass,
			config.SubscriptionId)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			fmt.Printf("Successfully deleted resource group '%s'.\n", resourceGroupName)
		}
		return
	}

	fmt.Printf("Creating resource group '%s'...\n", resourceGroupName)
	//Create a resource group on Azure Stack
	_, errRgStack := hybridresources.CreateResourceGroup(
		cntx,
		resourceGroupName,
		config.Location,
		config.CertPath,
		config.ResourceManagerUrl,
		config.TenantId,
		config.ClientId,
		config.CertPass,
		config.SubscriptionId)
	if errRgStack != nil {
		log.Fatal(errRgStack.Error())
	} else {
		fmt.Printf("Successfully created resource group '%s'.\n", resourceGroupName)
	}

	fmt.Printf("Getting storage account client...\n")
	// Create a storge account client
	storageAccountClient := hybridstorage.GetStorageAccountsClient(
		config.TenantId,
		config.ClientId,
		config.CertPass,
		config.ResourceManagerUrl,
		config.CertPath,
		config.SubscriptionId)
	if &storageAccountClient == nil {
		log.Fatal("Failed to get storage account client.\n")
	} else {
		fmt.Printf("Successfully got storage account client.\n")
	}

	fmt.Printf("Creating storage account '%s'...\n", storageAccountName)
	// Create storage account
	_, errSa := hybridstorage.CreateStorageAccount(
		cntx,
		storageAccountClient,
		storageAccountName,
		resourceGroupName,
		config.Location)
	if errSa != nil {
		log.Fatal(errSa.Error())
	} else {
		fmt.Printf("Successfully created storage account %s.\n", storageAccountName)
	}

	fmt.Printf("Getting dataplane URL...\n")
	dataplaneURL, errDP := hybridstorage.GetDataplaneURL(
		cntx,
		storageAccountClient,
		storageEndpointSuffix,
		storageAccountName,
		resourceGroupName,
		storageContainerName)
	if errDP != nil {
		log.Fatal(errDP.Error())
	} else {
		fmt.Printf("Successfully got dataplane URL.\n")
	}

	dirname, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory.\n")
	}
	blobFileAddress := dirname + filepath.FromSlash("/assets/test-upload-file.txt")
	fmt.Printf("Uploading '%s' to storage container...\n", blobFileAddress)
	uploadErr := hybridstorage.UploadDataToContainer(
		cntx,
		dataplaneURL,
		blobFileAddress)
	if uploadErr != nil {
		log.Fatal(uploadErr.Error())
	} else {
		fmt.Printf("Successfully uploaded '%s' to storage container...\n", blobFileAddress)
		fmt.Printf("Sample completed successfully.\n")
	}

	return
}
