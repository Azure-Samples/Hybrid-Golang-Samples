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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

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

const (
	publisher = "Canonical"
	offer     = "UbuntuServer"
	sku       = "16.04-LTS"
)

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

	var resourceGroupName = "TestGoVMSampleResourceGroup"

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

	fmt.Println("Creating a virtual network client")

	vnetClient, err := armnetwork.NewVirtualNetworksClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("\nError creating vnet client: %s\n", err)
		os.Exit(1)
	}

	//Create Vnet
	fmt.Println("Creating Vnet and subnets")
	var vnetName = "TestGoVnetName"
	var subnetName = "TestGoSubnetName"
	vnetresp, err := vnetClient.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		vnetName,
		armnetwork.VirtualNetwork{
			Location: to.Ptr(config.Location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{to.Ptr("10.0.0.0/8")},
				},
				Subnets: []*armnetwork.Subnet{
					to.Ptr(armnetwork.Subnet{
						Name: to.Ptr(subnetName),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.Ptr("10.0.0.0/16"),
						},
					}),
				},
			},
		},
		nil,
	)
	if err != nil {
		fmt.Printf("\nError creating Vnet: %s\n", err)
	}
	cntxTimeout, cancel := context.WithTimeout(cntx, 60*time.Second)
	defer cancel()
	_, err = vnetresp.PollUntilDone(cntxTimeout, nil)
	if err != nil {
		fmt.Printf("\nError creating Vnet: %s\n", err)
	}

	//Create NSG
	nsgName := "TestGoNsgName"
	nsgclient, err := armnetwork.NewSecurityGroupsClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("\nError creating NSG client: %s\n", err)
	}

	nsgresp, err := nsgclient.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		nsgName,
		armnetwork.SecurityGroup{
			Location: to.Ptr(config.Location),
			Properties: &armnetwork.SecurityGroupPropertiesFormat{
				SecurityRules: []*armnetwork.SecurityRule{
					&armnetwork.SecurityRule{
						Name: to.Ptr("allow_ssh"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
							SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
							SourcePortRange:          to.Ptr("1-65535"),
							DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
							DestinationPortRange:     to.Ptr("22"),
							Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
							Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
							Priority:                 to.Ptr(int32(100)),
						},
					},
					&armnetwork.SecurityRule{
						Name: to.Ptr("allow_https"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
							SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
							SourcePortRange:          to.Ptr("1-65535"),
							DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
							DestinationPortRange:     to.Ptr("443"),
							Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
							Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
							Priority:                 to.Ptr(int32(200)),
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		fmt.Printf("Failed to create nsg: %s\n", err)
		os.Exit(1)
	}
	defer cancel()
	_, err = nsgresp.PollUntilDone(cntxTimeout, nil)
	if err != nil {
		fmt.Printf("\nError creating nsg: %s\n", err)
	}

	// Create public ip
	fmt.Println("Creating public ip client")

	ipClient, err := armnetwork.NewPublicIPAddressesClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("Failed to create public ip client: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Creating public ip")
	var publicIpName = "TestGoIpAddr"
	ipresp, err := ipClient.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		publicIpName,
		armnetwork.PublicIPAddress{
			Name:     to.Ptr(publicIpName),
			Location: to.Ptr(config.Location),
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
			},
		},
		nil,
	)
	if err != nil {
		fmt.Printf("Failed to create public ip: %s\n", err)
		os.Exit(1)
	}
	defer cancel()
	_, err = ipresp.PollUntilDone(cntxTimeout, nil)
	if err != nil {
		fmt.Printf("\nError creating public ip: %s\n", err)
	}

	//Get subnet
	fmt.Println("Create Subnet client")
	subnetClient, err := armnetwork.NewSubnetsClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("Failed to create subnets client: %s\n", err)
		os.Exit(1)
	}

	subresp, err := subnetClient.Get(context.Background(), resourceGroupName, vnetName, subnetName, nil)
	if err != nil {
		fmt.Printf("Failed to get subnet: %s\n", err)
		os.Exit(1)
	}

	//Create a network interface
	fmt.Println("Creating a Network Interface client")
	niClient, err := armnetwork.NewInterfacesClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("Failed to create network interface client: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Creating Network Interface")
	var nicname = "testGoNetworkInterface"
	nsg, _ := nsgresp.Result(context.Background())
	pubIp, _ := ipresp.Result(context.Background())
	nicresp, err := niClient.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		nicname,
		armnetwork.Interface{
			Name:     &nicname,
			Location: to.Ptr(config.Location),
			Properties: &armnetwork.InterfacePropertiesFormat{
				NetworkSecurityGroup: &nsg.SecurityGroup,
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					&armnetwork.InterfaceIPConfiguration{
						Name: to.Ptr("ipConfig1"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							Subnet:                    &subresp.Subnet,
							PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
							PublicIPAddress:           &pubIp.PublicIPAddress,
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		fmt.Printf("Failed to create network interface: %s\n", err)
		os.Exit(1)
	}
	defer cancel()
	_, err = nicresp.PollUntilDone(cntxTimeout, nil)
	if err != nil {
		fmt.Printf("\nError creating network interface: %s\n", err)
	}
	nicresult, _ := nicresp.Result(context.Background())
	nic := nicresult.Interface

	// Create storage acc
	var storageAccountName = "govmteststorageacc"
	saClient, err := armstorage.NewAccountsClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("\nErr creating storage client %s", err)
	}

	var skuname = armstorage.SKUNameStandardLRS

	_, err = saClient.BeginCreate(
		context.Background(),
		resourceGroupName,
		storageAccountName,
		armstorage.AccountCreateParameters{
			SKU:        &armstorage.SKU{Name: &skuname},
			Location:   to.Ptr(config.Location),
			Properties: &armstorage.AccountPropertiesCreateParameters{},
		},
		nil)

	if err != nil {
		fmt.Printf("\nErr creating storage account: %s", err)
		os.Exit(1)
	}

	// Create Virtual Machine
	var vmName = "TestGoVm1"
	fmt.Println("Creating Virtual Machine client")

	vmClient, err := armcompute.NewVirtualMachinesClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("\nErr creating vm client: %s", err)
		os.Exit(1)
	}

	// Create Profiles
	hardwareProfile := &armcompute.HardwareProfile{
		VMSize: to.Ptr(armcompute.VirtualMachineSizeTypesStandardA1),
	}

	vhdURItemplate := "https://%s.blob." + environment.StorageEndpointSuffix + "/vhds/%s.vhd"
	storageProfile := &armcompute.StorageProfile{
		ImageReference: &armcompute.ImageReference{
			Publisher: to.Ptr(publisher),
			Offer:     to.Ptr(offer),
			SKU:       to.Ptr(sku),
			Version:   to.Ptr("latest"),
		},
		OSDisk: &armcompute.OSDisk{
			Name: to.Ptr("osDisk"),
			Vhd: &armcompute.VirtualHardDisk{
				URI: to.Ptr(fmt.Sprintf(vhdURItemplate, storageAccountName, vmName)),
			},
			CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
		},
	}

	osProfile := &armcompute.OSProfile{
		ComputerName:  to.Ptr(vmName),
		AdminUsername: to.Ptr("username"),
		AdminPassword: to.Ptr("Password!23"),
	}

	networkProfile := &armcompute.NetworkProfile{
		NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
			&armcompute.NetworkInterfaceReference{
				ID: nic.ID,
				Properties: &armcompute.NetworkInterfaceReferenceProperties{
					Primary: to.Ptr(true),
				},
			},
		},
	}

	fmt.Println("Creating Virtual Machine")
	_, err = vmClient.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		vmName,
		armcompute.VirtualMachine{
			Location: to.Ptr(config.Location),
			Properties: &armcompute.VirtualMachineProperties{
				HardwareProfile: hardwareProfile,
				OSProfile:       osProfile,
				NetworkProfile:  networkProfile,
				StorageProfile:  storageProfile,
			},
		},
		nil,
	)
	if err != nil {
		fmt.Printf("\nErr creating vm: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Listing virtual machines in %s\n", resourceGroupName)
	pager := vmClient.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			fmt.Printf("\nErr can't get next page in vm list")
			os.Exit(1)
		}
		if resp.VirtualMachineListResult.Value != nil {
			for _, vm := range resp.VirtualMachineListResult.Value {
				fmt.Print(*vm.Name + ", ")
			}
		}
	}
	fmt.Println()

	fmt.Println("Deleting VM")
	delResp, _ := vmClient.BeginDelete(context.Background(), resourceGroupName, vmName, nil)
	cntxTimeoutDel, cancel := context.WithTimeout(cntx, 500*time.Second)
	defer cancel()
	_, err = delResp.PollUntilDone(cntxTimeoutDel, nil)
	if err != nil {
		fmt.Printf("\nError deleting vm: %s\n", err)
	}

	//Managed disk vm
	fmt.Println("Creating Disk client")
	diskClient, err := armcompute.NewDisksClient(config.SubscriptionId, cred, &rgoptions)
	if err != nil {
		fmt.Printf("\nErr creating disk client: %s", err)
		os.Exit(1)
	}
	var diskName = "osDisk2"
	var vmNameMD = "TestGoManagedDiskVm"
	fmt.Println("Creating Disk")
	diskResp, _ := diskClient.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		diskName,
		armcompute.Disk{
			Location: to.Ptr(config.Location),
			Properties: &armcompute.DiskProperties{
				CreationData: &armcompute.CreationData{
					CreateOption: to.Ptr(armcompute.DiskCreateOptionEmpty),
				},
				DiskSizeGB: to.Ptr(int32(1)),
			},
		},
		nil,
	)
	if err != nil {
		fmt.Printf("\nErr creating disk: %s", err)
		os.Exit(1)
	}
	defer cancel()
	_, err = diskResp.PollUntilDone(cntxTimeout, nil)
	if err != nil {
		fmt.Printf("\nError creating disk: %s\n", err)
	}
	diskresult, _ := diskResp.Result(context.Background())
	disk := diskresult.Disk

	storageProfileManagedDisk := &armcompute.StorageProfile{
		ImageReference: &armcompute.ImageReference{
			Publisher: to.Ptr(publisher),
			Offer:     to.Ptr(offer),
			SKU:       to.Ptr(sku),
			Version:   to.Ptr("latest"),
		},
		DataDisks: []*armcompute.DataDisk{
			&armcompute.DataDisk{
				CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesAttach),
				ManagedDisk: &armcompute.ManagedDiskParameters{
					StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS),
					ID:                 disk.ID,
				},
				Caching:    to.Ptr(armcompute.CachingTypesReadOnly),
				DiskSizeGB: to.Ptr(int32(1)),
				Lun:        to.Ptr(int32(1)),
				Name:       to.Ptr(diskName),
			},
		},
		OSDisk: &armcompute.OSDisk{
			Name:         to.Ptr("osDiskMD"),
			CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
		},
	}

	fmt.Println("Creating Managed Disk VM")
	_, err = vmClient.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		vmNameMD,
		armcompute.VirtualMachine{
			Location: to.Ptr(config.Location),
			Properties: &armcompute.VirtualMachineProperties{
				HardwareProfile: hardwareProfile,
				OSProfile:       osProfile,
				NetworkProfile:  networkProfile,
				StorageProfile:  storageProfileManagedDisk,
			},
		},
		nil,
	)
	if err != nil {
		fmt.Printf("\nErr creating managed disk vm: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Listing virtual machines in %s\n", resourceGroupName)
	pager = vmClient.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			fmt.Printf("\nErr can't get next page in vm list")
			os.Exit(1)
		}
		if resp.VirtualMachineListResult.Value != nil {
			for _, vm := range resp.VirtualMachineListResult.Value {
				fmt.Print(*vm.Name + ", ")
			}
		}
	}
	fmt.Println()

	if *clean {
		fmt.Println("Deleting resource group")
		result, err := rgClient.BeginDelete(context.Background(), resourceGroupName, nil)
		if err != nil {
			fmt.Printf("Failed to delete resource group: %s\n", resourceGroupName)
			os.Exit(1)
		}

		cntxTimeout, cancel = context.WithTimeout(cntx, 500*time.Second)
		defer cancel()
		_, err = result.PollUntilDone(cntxTimeout, nil)
		if err != nil {
			fmt.Println("Timed out when deleting resource group")
			os.Exit(1)
		}
	}
}
