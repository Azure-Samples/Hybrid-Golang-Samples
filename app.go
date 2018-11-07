package main

import (
	"context"
	"fmt"

	hybridresources "hybridSample/hybridResources"
	hybridstorage "hybridSample/hybridStorage"
	"hybridSample/hybridcompute"
	"hybridSample/hybridnetwork"
	"os"
	"strings"
)

var (
	stackActiveDirectoryEndpoint   = os.Getenv("ACTIVE_DIRECTORY_ENDPOINT")
	stackActiveDirectoryResourceID = os.Getenv("ACTIVE_DIRECTORY_RESOURCE_ID")
	stackArmEndpoint               = os.Getenv("ARM_ENDPOINT")
	stackTenantID                  = os.Getenv("AZURE_TENANT_ID")
	stackClientID                  = os.Getenv("AZURE_CLIENT_ID")
	stackClientSecret              = os.Getenv("AZURE_CLIENT_SECRET")
	stackSubscriptionID            = os.Getenv("AZURE_SUBSCRIPTION_ID")
	stackStorageEndpointSuffix     = os.Getenv("AZURE_STORAGE_ENDPOINT_SUFFIX")
	stackLocation                  = os.Getenv("AZURE_LOCATION")

	vmName             = "az-samples-go-vmname"
	nicName            = "nic1"
	username           = "az-samples-go-user"
	password           = "NoSoupForYou1!"
	sshPublicKeyPath   = os.Getenv("HOME") + "/.ssh/id_rsa.pub"
	virtualNetworkName = "vnet1"
	subnetName         = "subnet1"
	nsgName            = "nsg1"
	ipName             = "ip1"
	storageAccountName = strings.ToLower("samplestacc123")
	stackRgName        = "stackrg"
)

func main() {
	cntx := context.Background()

	//Create a resource group on Azure Stack
	_, errRgStack := hybridresources.CreateResourceGroup(
		cntx,
		stackRgName,
		"stack",
		stackLocation,
		stackActiveDirectoryEndpoint,
		stackActiveDirectoryResourceID,
		stackArmEndpoint,
		stackTenantID,
		stackClientID,
		stackClientSecret,
		stackSubscriptionID)
	if errRgStack != nil {
		fmt.Println(errRgStack.Error())
		return
	}

	// Create virtual network on Azure Stack
	_, errVnet := hybridnetwork.CreateVirtualNetworkAndSubnets(
		cntx,
		virtualNetworkName,
		subnetName,
		stackActiveDirectoryEndpoint,
		stackTenantID,
		stackClientID,
		stackClientSecret,
		stackActiveDirectoryResourceID,
		stackArmEndpoint,
		stackSubscriptionID,
		stackRgName,
		stackLocation)
	if errVnet != nil {
		fmt.Println(errVnet.Error())
		return
	}

	//Create network security group on Azure Stack
	_, errSg := hybridnetwork.CreateNetworkSecurityGroup(
		cntx,
		nsgName,
		stackActiveDirectoryEndpoint,
		stackTenantID,
		stackClientID,
		stackClientSecret,
		stackActiveDirectoryResourceID,
		stackArmEndpoint,
		stackSubscriptionID,
		stackRgName,
		stackLocation)
	if errSg != nil {
		fmt.Println(errSg.Error())
		return
	}

	// Create public IP on Azure Stack
	_, errIP := hybridnetwork.CreatePublicIP(
		cntx,
		ipName,
		stackActiveDirectoryEndpoint,
		stackTenantID,
		stackClientID,
		stackClientSecret,
		stackActiveDirectoryResourceID,
		stackArmEndpoint,
		stackSubscriptionID,
		stackRgName,
		stackLocation)
	if errIP != nil {
		fmt.Println(errIP.Error())
	}

	// Create network interface on Azure Stack
	_, errNic := hybridnetwork.CreateNetworkInterface(
		cntx,
		nicName,
		nsgName,
		virtualNetworkName,
		subnetName,
		ipName,
		stackActiveDirectoryEndpoint,
		stackTenantID,
		stackClientID,
		stackClientSecret,
		stackActiveDirectoryResourceID,
		stackArmEndpoint,
		stackSubscriptionID,
		stackRgName,
		stackLocation)
	if errNic != nil {
		fmt.Println(errNic.Error())
	}

	// Create storage account and disk on Azure Stack
	_, errSa := hybridstorage.CreateStorageAccount(
		cntx,
		storageAccountName,
		stackRgName,
		stackLocation,
		stackActiveDirectoryEndpoint,
		stackTenantID,
		stackClientID,
		stackClientSecret,
		stackActiveDirectoryResourceID,
		stackArmEndpoint,
		stackSubscriptionID)
	if errSa != nil {
		fmt.Println(errSa.Error())
	}

	// Create virtual machine on Azure Stack
	_, errVM := hybridcompute.CreateVM(cntx,
		vmName,
		nicName,
		username,
		password,
		storageAccountName,
		sshPublicKeyPath,
		stackRgName,
		stackLocation,
		stackActiveDirectoryEndpoint,
		stackTenantID,
		stackClientID,
		stackClientSecret,
		stackActiveDirectoryResourceID,
		stackArmEndpoint,
		stackSubscriptionID,
		stackStorageEndpointSuffix)
	if errVM != nil {
		fmt.Println(errVM.Error())
	}
	return
}
