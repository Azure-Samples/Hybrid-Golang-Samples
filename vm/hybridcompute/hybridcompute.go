// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package hybridcompute

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"vm/hybridnetwork"
	"vm/iam"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	publisher   = "Canonical"
	offer       = "UbuntuServer"
	sku         = "16.04-LTS"
	errorPrefix = "Cannot create VM, reason: %v"
)

func getVMClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID string) compute.VirtualMachinesClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	vmClient := compute.NewVirtualMachinesClientWithBaseURI(armEndpoint, subscriptionID)
	vmClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return vmClient
}

// CreateVM creates a new virtual machine with the specified name using the specified network interface and storage account.
// Username, password, and sshPublicKeyPath determine logon credentials.
func CreateVM(ctx context.Context, vmName, nicName, username, password, storageAccountName, sshPublicKeyPath, rgName, location, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, storageEndpointSuffix string) (vm compute.VirtualMachine, err error) {
	cntx := context.Background()
	nic, _ := hybridnetwork.GetNic(cntx, nicName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName)

	var sshKeyData string
	_, err = os.Stat(sshPublicKeyPath)
	if err == nil {
		sshBytes, err := ioutil.ReadFile(sshPublicKeyPath)
		if err != nil {
			log.Fatalf(fmt.Sprintf(errorPrefix, fmt.Sprintf("failed to read SSH key data: %v", err)))
		}
		sshKeyData = string(sshBytes)
	}

	vhdURItemplate := "https://%s.blob." + storageEndpointSuffix + "/vhds/%s.vhd"
	vmClient := getVMClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	hardwareProfile := &compute.HardwareProfile{
		VMSize: compute.StandardA1,
	}
	storageProfile := &compute.StorageProfile{
		ImageReference: &compute.ImageReference{
			Publisher: to.StringPtr(publisher),
			Offer:     to.StringPtr(offer),
			Sku:       to.StringPtr(sku),
			Version:   to.StringPtr("latest"),
		},
		OsDisk: &compute.OSDisk{
			Name: to.StringPtr("osDisk"),
			Vhd: &compute.VirtualHardDisk{
				URI: to.StringPtr(fmt.Sprintf(vhdURItemplate, storageAccountName, vmName)),
			},
			CreateOption: compute.DiskCreateOptionTypesFromImage,
		},
	}
	var osProfile *compute.OSProfile
	if len(username) != 0 && len(sshKeyData) != 0 {
		osProfile = &compute.OSProfile{
			ComputerName:  to.StringPtr(vmName),
			AdminUsername: to.StringPtr(username),
			LinuxConfiguration: &compute.LinuxConfiguration{
				SSH: &compute.SSHConfiguration{
					PublicKeys: &[]compute.SSHPublicKey{
						{
							Path:    to.StringPtr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", username)),
							KeyData: to.StringPtr(sshKeyData),
						},
					},
				},
			},
		}
	} else if len(username) != 0 && len(password) != 0 {
		osProfile = &compute.OSProfile{
			ComputerName:  to.StringPtr(vmName),
			AdminUsername: to.StringPtr(username),
			AdminPassword: to.StringPtr(password),
		}
	} else if len(sshKeyData) == 0 && len(password) == 0 {
		log.Fatalf(fmt.Sprintf(errorPrefix, fmt.Sprintf("Both VM admin password and SSH key pair path %s are invalid. At least one required to create VM. Usage for password authentication: go run app.go <PASSWORD>", sshPublicKeyPath)))
	} else {
		log.Fatalf(fmt.Sprintf(errorPrefix, fmt.Sprintf("VM admin username is an empty string.")))
	}
	networkProfile := &compute.NetworkProfile{
		NetworkInterfaces: &[]compute.NetworkInterfaceReference{
			{
				ID: nic.ID,
				NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
					Primary: to.BoolPtr(true),
				},
			},
		},
	}
	virtualMachine := compute.VirtualMachine{
		Location: to.StringPtr(location),
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: hardwareProfile,
			StorageProfile:  storageProfile,
			OsProfile:       osProfile,
			NetworkProfile:  networkProfile,
		},
	}
	future, err := vmClient.CreateOrUpdate(
		cntx,
		rgName,
		vmName,
		virtualMachine,
	)
	if err != nil {
		return vm, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletionRef(cntx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	return future.Result(vmClient)
}
