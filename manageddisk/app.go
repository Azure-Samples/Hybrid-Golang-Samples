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

	hybridresources "manageddisk/hybridResources"
	hybridstorage "manageddisk/hybridStorage"
	"manageddisk/hybridcompute"
	"manageddisk/hybridnetwork"

	"github.com/Azure/go-autorest/autorest/azure"
)

var (
	vmName             = "az-samples-go-vmname"
	nicName            = "nic1"
	username           = "VMAdmin"
	virtualNetworkName = "vnet1"
	subnetName         = "subnet1"
	nsgName            = "nsg1"
	ipName             = "ip1"
	storageAccountName = strings.ToLower("disksamplestacc")
	resourceGroupName  = "azure-sample-golang-manageddisk"
	diskName           = "sampledisk"
)

type AzureCertSpConfig struct {
	ClientId                   string
	CertPass                   string
	CertPath                   string
	ObjectId                   string
	SubscriptionId             string
	TenantId                   string
	ResourceManagerEndpointUrl string
	Location                   string
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

	// Password is not required when using SSH key pair.
	var password string
	if len(os.Args) == 2 {
		password = os.Args[1]
	} else if len(os.Args) > 2 {
		log.Fatalf("Error, invalid number of CLI arguments: %d", len(os.Args))
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not find user home directory. The sample code looks for .ssh folder in the user home directory %s.", homeDir)
	}
	sshPublicKeyPath := homeDir + filepath.FromSlash("/.ssh/id_rsa.pub")
	_, sshPubFileErr := os.Stat(sshPublicKeyPath)
	if sshPubFileErr != nil && len(os.Args) == 1 {
		log.Fatalf("Both VM admin password and SSH key pair path %s are invalid. At least one required to create VM. Usage for password authentication: go run app.go <PASSWORD>", sshPublicKeyPath)
	}
	cntx := context.Background()
	environment, _ := azure.EnvironmentFromURL(config.ResourceManagerEndpointUrl)
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
			config.ResourceManagerEndpointUrl,
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
		config.ResourceManagerEndpointUrl,
		config.TenantId,
		config.ClientId,
		config.CertPass,
		config.SubscriptionId)
	if errRgStack != nil {
		log.Fatal(errRgStack.Error())
	} else {
		fmt.Printf("Successfully created resource group '%s'.\n", resourceGroupName)
	}

	fmt.Printf("Creating virtual network '%s' and subnet '%s'...\n", virtualNetworkName, subnetName)
	// Create virtual network on Azure Stack
	_, errVnet := hybridnetwork.CreateVirtualNetworkAndSubnets(
		cntx,
		virtualNetworkName,
		subnetName,
		config.CertPath,
		config.TenantId,
		config.ClientId,
		config.CertPass,
		config.ResourceManagerEndpointUrl,
		config.SubscriptionId,
		resourceGroupName,
		config.Location)
	if errVnet != nil {
		log.Fatal(errVnet.Error())
	} else {
		fmt.Printf("Successfully created virtual network '%s' and subnet '%s'.\n", virtualNetworkName, subnetName)
	}

	fmt.Printf("Creating network security group '%s'...\n", nsgName)
	//Create network security group on Azure Stack
	_, errSg := hybridnetwork.CreateNetworkSecurityGroup(
		cntx,
		nsgName,
		config.CertPath,
		config.TenantId,
		config.ClientId,
		config.CertPass,
		config.ResourceManagerEndpointUrl,
		config.SubscriptionId,
		resourceGroupName,
		config.Location)
	if errSg != nil {
		log.Fatal(errSg.Error())
	} else {
		fmt.Printf("Successfully created network security group '%s'.\n", nsgName)
	}

	fmt.Printf("Creating public ip '%s'...\n", ipName)
	// Create public IP on Azure Stack
	_, errIP := hybridnetwork.CreatePublicIP(
		cntx,
		ipName,
		config.CertPath,
		config.TenantId,
		config.ClientId,
		config.CertPass,
		config.ResourceManagerEndpointUrl,
		config.SubscriptionId,
		resourceGroupName,
		config.Location)
	if errIP != nil {
		log.Fatal(errIP.Error())
	} else {
		fmt.Printf("Successfully created public ip '%s'.\n", ipName)
	}

	fmt.Printf("Creating network interface '%s'...\n", nicName)
	// Create network interface on Azure Stack
	_, errNic := hybridnetwork.CreateNetworkInterface(
		cntx,
		nicName,
		nsgName,
		virtualNetworkName,
		subnetName,
		ipName,
		config.CertPath,
		config.TenantId,
		config.ClientId,
		config.CertPass,
		config.ResourceManagerEndpointUrl,
		config.SubscriptionId,
		resourceGroupName,
		config.Location)
	if errNic != nil {
		log.Fatal(errNic.Error())
	} else {
		fmt.Printf("Successfully created network interface '%s'.\n", nicName)
	}

	fmt.Printf("Creating storage account '%s'...\n", storageAccountName)
	// Create storage account and disk on Azure Stack
	_, errSa := hybridstorage.CreateStorageAccount(
		cntx,
		storageAccountName,
		resourceGroupName,
		config.Location,
		config.CertPath,
		config.TenantId,
		config.ClientId,
		config.CertPass,
		config.ResourceManagerEndpointUrl,
		config.SubscriptionId)
	if errSa != nil {
		log.Fatal(errSa.Error())
	} else {
		fmt.Printf("Successfully created storage account '%s'.\n", storageAccountName)
	}

	fmt.Printf("Creating vm '%s'...\n", vmName)
	// Create virtual machine on Azure Stack
	_, errVM := hybridcompute.CreateVM(cntx,
		vmName,
		diskName,
		nicName,
		username,
		password,
		storageAccountName,
		sshPublicKeyPath,
		resourceGroupName,
		config.Location,
		config.TenantId,
		config.ClientId,
		config.CertPass,
		config.CertPath,
		config.ResourceManagerEndpointUrl,
		config.SubscriptionId)
	if errVM != nil {
		log.Fatal(errVM.Error())
	} else {
		fmt.Printf("Successfully created vm '%s'.\n", vmName)
		fmt.Printf("Sample completed successfully.\n")
	}

	return
}
