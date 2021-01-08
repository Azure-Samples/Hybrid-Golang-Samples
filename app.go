package main

import (
    "context"
    "fmt"
    "os"
    "strings"

    hybridresources "Hybrid-Compute-Go-Create-VM/hybridResources"
    hybridstorage "Hybrid-Compute-Go-Create-VM/hybridStorage"
    "Hybrid-Compute-Go-Create-VM/hybridcompute"
    "Hybrid-Compute-Go-Create-VM/hybridnetwork"

    "github.com/Azure/go-autorest/autorest/azure"
)

var (
    armEndpoint    = os.Getenv("AZS_ARM_ENDPOINT")
    tenantID       = os.Getenv("AZS_TENANT_ID")
    clientID       = os.Getenv("AZS_SECRET_CLIENT_ID")
    clientSecret   = os.Getenv("AZS_CLIENT_SECRET")
    subscriptionID = os.Getenv("AZS_SUBSCRIPTION_ID")
    location       = os.Getenv("AZS_LOCATION")

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
    rgName             = "stackrg"
)

func main() {
    cntx := context.Background()
    environment, _ := azure.EnvironmentFromURL(armEndpoint)
    storageEndpointSuffix := environment.StorageEndpointSuffix
    //Create a resource group on Azure Stack
    _, errRgStack := hybridresources.CreateResourceGroup(
        cntx,
        rgName,
        location,
        armEndpoint,
        tenantID,
        clientID,
        clientSecret,
        subscriptionID)
    if errRgStack != nil {
        fmt.Println(errRgStack.Error())
        return
    }

    // Create virtual network on Azure Stack
    _, errVnet := hybridnetwork.CreateVirtualNetworkAndSubnets(
        cntx,
        virtualNetworkName,
        subnetName,
        tenantID,
        clientID,
        clientSecret,
        armEndpoint,
        subscriptionID,
        rgName,
        location)
    if errVnet != nil {
        fmt.Println(errVnet.Error())
        return
    }

    //Create network security group on Azure Stack
    _, errSg := hybridnetwork.CreateNetworkSecurityGroup(
        cntx,
        nsgName,
        tenantID,
        clientID,
        clientSecret,
        armEndpoint,
        subscriptionID,
        rgName,
        location)
    if errSg != nil {
        fmt.Println(errSg.Error())
        return
    }

    // Create public IP on Azure Stack
    _, errIP := hybridnetwork.CreatePublicIP(
        cntx,
        ipName,
        tenantID,
        clientID,
        clientSecret,
        armEndpoint,
        subscriptionID,
        rgName,
        location)
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
        tenantID,
        clientID,
        clientSecret,
        armEndpoint,
        subscriptionID,
        rgName,
        location)
    if errNic != nil {
        fmt.Println(errNic.Error())
    }

    // Create storage account and disk on Azure Stack
    _, errSa := hybridstorage.CreateStorageAccount(
        cntx,
        storageAccountName,
        rgName,
        location,
        tenantID,
        clientID,
        clientSecret,
        armEndpoint,
        subscriptionID)
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
        rgName,
        location,
        tenantID,
        clientID,
        clientSecret,
        armEndpoint,
        subscriptionID,
        storageEndpointSuffix)
    if errVM != nil {
        fmt.Println(errVM.Error())
    }
    return
}
