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

	"manageddisk/hybridnetwork"
	"manageddisk/iam"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

const (
	publisher   = "Canonical"
	offer       = "UbuntuServer"
	sku         = "16.04-LTS"
	errorPrefix = "Cannot create VM, reason: %v"
)

func getVMClient(certPath, tenantID, clientID, certPass, subscriptionID string) (*armcompute.VirtualMachinesClient, error) {
	token, err := iam.GetResourceManagementToken(tenantID, clientID, certPass, certPath)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	return armcompute.NewVirtualMachinesClient(subscriptionID, token, nil)
}

func getDiskClient(certPath, tenantID, clientID, certPass, subscriptionID string) (*armcompute.DisksClient, error) {
	token, err := iam.GetResourceManagementToken(tenantID, clientID, certPass, certPath)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	return armcompute.NewDisksClient(subscriptionID, token, nil)
}

// CreateVM creates a new virtual machine with the specified name using the specified network interface and storage account.
// Username, password, and sshPublicKeyPath determine logon credentials.
func CreateVM(ctx context.Context, vmName, diskName, nicName, username, password, storageAccountName, sshPublicKeyPath, rgName, location, tenantID, clientID, certPass, certPath, subscriptionID string) (vm armcompute.VirtualMachine, err error) {
	cntx := context.Background()
	nic, _ := hybridnetwork.GetNic(cntx, nicName, certPath, tenantID, clientID, certPass, subscriptionID, rgName)

	var sshKeyData string
	_, err = os.Stat(sshPublicKeyPath)
	if err == nil {
		sshBytes, err := ioutil.ReadFile(sshPublicKeyPath)
		if err != nil {
			log.Fatalf(fmt.Sprintf(errorPrefix, fmt.Sprintf("failed to read SSH key data: %v", err)))
		}
		sshKeyData = string(sshBytes)
	}

	vmClient, err := getVMClient(certPath, tenantID, clientID, certPass, subscriptionID)
	if err != nil {
		return vm, err
	}
	diskClient, err := getDiskClient(certPath, tenantID, clientID, certPass, subscriptionID)
	if err != nil {
		return vm, err
	}
	diskFuture, _ := diskClient.BeginCreateOrUpdate(
		ctx,
		rgName,
		diskName,
		armcompute.Disk{
			Location: to.Ptr(location),
			Properties: &armcompute.DiskProperties{
				CreationData: &armcompute.CreationData{
					CreateOption: to.Ptr(armcompute.DiskCreateOptionEmpty),
				},
				DiskSizeGB: to.Ptr[int32](1),
			},
		},
		nil,
	)
	diskResp, err := diskFuture.PollUntilDone(cntx, nil)
	if err != nil {
		return vm, err
	}
	disk := diskResp.Disk
	hardwareProfile := &armcompute.HardwareProfile{
		VMSize: to.Ptr(armcompute.VirtualMachineSizeTypesStandardA1),
	}
	storageProfile := &armcompute.StorageProfile{
		ImageReference: &armcompute.ImageReference{
			Publisher: to.Ptr(publisher),
			Offer:     to.Ptr(offer),
			SKU:       to.Ptr(sku),
			Version:   to.Ptr("latest"),
		},
		DataDisks: []*armcompute.DataDisk{
			{
				CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesAttach),
				ManagedDisk: &armcompute.ManagedDiskParameters{
					StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS),
					ID:                 disk.ID,
				},
				Caching:    to.Ptr(armcompute.CachingTypesReadOnly),
				DiskSizeGB: to.Ptr[int32](1),
				Lun:        to.Ptr[int32](1),
				Name:       to.Ptr(diskName),
			},
		},
		OSDisk: &armcompute.OSDisk{
			Name:         to.Ptr("osDisk"),
			CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
		},
	}
	var osProfile *armcompute.OSProfile
	if len(username) != 0 && len(sshKeyData) != 0 {
		osProfile = &armcompute.OSProfile{
			ComputerName:  to.Ptr(vmName),
			AdminUsername: to.Ptr(username),
			LinuxConfiguration: &armcompute.LinuxConfiguration{
				SSH: &armcompute.SSHConfiguration{
					PublicKeys: []*armcompute.SSHPublicKey{
						{
							Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", username)),
							KeyData: to.Ptr(sshKeyData),
						},
					},
				},
			},
		}
	} else if len(username) != 0 && len(password) != 0 {
		osProfile = &armcompute.OSProfile{
			ComputerName:  to.Ptr(vmName),
			AdminUsername: to.Ptr(username),
			AdminPassword: to.Ptr(password),
		}
	} else if len(sshKeyData) == 0 && len(password) == 0 {
		log.Fatalf(fmt.Sprintf(errorPrefix, fmt.Sprintf("Both VM admin password and SSH key pair path %s are invalid. At least one required to create VM. Usage for password authentication: go run app.go <PASSWORD>", sshPublicKeyPath)))
	} else {
		log.Fatalf(fmt.Sprintf(errorPrefix, fmt.Sprintf("VM admin username is an empty string.")))
	}
	networkProfile := &armcompute.NetworkProfile{
		NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
			{
				ID: nic.ID,
				Properties: &armcompute.NetworkInterfaceReferenceProperties{
					Primary: to.Ptr(true),
				},
			},
		},
	}
	virtualMachine := armcompute.VirtualMachine{
		Location: to.Ptr(location),
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: hardwareProfile,
			StorageProfile:  storageProfile,
			OSProfile:       osProfile,
			NetworkProfile:  networkProfile,
		},
	}
	future, err := vmClient.BeginCreateOrUpdate(
		cntx,
		rgName,
		vmName,
		virtualMachine,
		nil,
	)
	if err != nil {
		return vm, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	resp, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return vm, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	return resp.VirtualMachine, nil
}
