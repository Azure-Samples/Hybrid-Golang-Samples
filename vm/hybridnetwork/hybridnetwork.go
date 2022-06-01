// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package hybridnetwork

import (
	"context"
	"fmt"
	"log"

	"vm/iam"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

const (
	errorPrefix = "Cannot create %v, reason: %v"
)

func getVnetClient(tenantID, clientID, clientSecret, subscriptionID string) (*armnetwork.VirtualNetworksClient, error) {
	token, err := iam.GetResourceManagementTokenHybrid(tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "virtual network", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	return armnetwork.NewVirtualNetworksClient(subscriptionID, token, nil)
}

func getNsgClient(tenantID, clientID, clientSecret, subscriptionID string) (*armnetwork.SecurityGroupsClient, error) {
	token, err := iam.GetResourceManagementTokenHybrid(tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "security group", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	return armnetwork.NewSecurityGroupsClient(subscriptionID, token, nil)
}

func getIPClient(tenantID, clientID, clientSecret, subscriptionID string) (*armnetwork.PublicIPAddressesClient, error) {
	token, err := iam.GetResourceManagementTokenHybrid(tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "public IP address", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	return armnetwork.NewPublicIPAddressesClient(subscriptionID, token, nil)
}

func getNicClient(tenantID, clientID, clientSecret, subscriptionID string) (*armnetwork.InterfacesClient, error) {
	token, err := iam.GetResourceManagementTokenHybrid(tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "network interface", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	return armnetwork.NewInterfacesClient(subscriptionID, token, nil)
}

func getSubnetsClient(tenantID, clientID, clientSecret, subscriptionID string) (*armnetwork.SubnetsClient, error) {
	token, err := iam.GetResourceManagementTokenHybrid(tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "subnet", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	return armnetwork.NewSubnetsClient(subscriptionID, token, nil)
}

// CreateVirtualNetworkAndSubnets creates a virtual network with two subnets
func CreateVirtualNetworkAndSubnets(cntx context.Context, vnetName, subnetName, tenantID, clientID, clientSecret, subscriptionID, rgName, location string) (vnet armnetwork.VirtualNetwork, err error) {
	resourceName := "virtual network and subnet"
	vnetClient, err := getVnetClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return vnet, err
	}
	future, err := vnetClient.BeginCreateOrUpdate(
		cntx,
		rgName,
		vnetName,
		armnetwork.VirtualNetwork{
			Location: to.Ptr(location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{to.Ptr("10.0.0.0/8")},
				},
				Subnets: []*armnetwork.Subnet{
					{
						Name: to.Ptr(subnetName),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.Ptr("10.0.0.0/16"),
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return vnet, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	resp, err := future.PollUntilDone(cntx, nil)
	if err != nil {
		return vnet, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get the vnet create or update future response: %v", err)))
	}

	return resp.VirtualNetwork, nil
}

// CreateNetworkSecurityGroup creates a new network security group
func CreateNetworkSecurityGroup(cntx context.Context, nsgName, tenantID, clientID, clientSecret, subscriptionID, rgName, location string) (nsg armnetwork.SecurityGroup, err error) {
	resourceName := "security group"
	nsgClient, err := getNsgClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return nsg, err
	}
	future, err := nsgClient.BeginCreateOrUpdate(
		cntx,
		rgName,
		nsgName,
		armnetwork.SecurityGroup{
			Location: to.Ptr(location),
			Properties: &armnetwork.SecurityGroupPropertiesFormat{
				SecurityRules: []*armnetwork.SecurityRule{
					{
						Name: to.Ptr("allow_ssh"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
							SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
							SourcePortRange:          to.Ptr("1-65535"),
							DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
							DestinationPortRange:     to.Ptr("22"),
							Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
							Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
							Priority:                 to.Ptr[int32](100),
						},
					},
					{
						Name: to.Ptr("allow_https"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
							SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
							SourcePortRange:          to.Ptr("1-65535"),
							DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
							DestinationPortRange:     to.Ptr("443"),
							Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
							Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
							Priority:                 to.Ptr[int32](200),
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nsg, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	resp, err := future.PollUntilDone(cntx, nil)
	if err != nil {
		return nsg, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get nsg create or update future response: %v", err)))
	}

	return resp.SecurityGroup, nil
}

// CreatePublicIP creates a new public IP
func CreatePublicIP(cntx context.Context, ipName, tenantID, clientID, clientSecret, subscriptionID, rgName, location string) (ip armnetwork.PublicIPAddress, err error) {
	resourceName := "public IP"
	ipClient, err := getIPClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return ip, err
	}
	future, err := ipClient.BeginCreateOrUpdate(
		cntx,
		rgName,
		ipName,
		armnetwork.PublicIPAddress{
			Name:     to.Ptr(ipName),
			Location: to.Ptr(location),
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
			},
		},
		nil,
	)
	if err != nil {
		return ip, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	resp, err := future.PollUntilDone(cntx, nil)
	if err != nil {
		return ip, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get public ip address create or update future response: %v", err)))
	}
	return resp.PublicIPAddress, nil
}

// CreateNetworkInterface creates a new network interface
func CreateNetworkInterface(cntx context.Context, netInterfaceName, nsgName, vnetName, subnetName, ipName, tenantID, clientID, clientSecret, subscriptionID, rgName, location string) (nic armnetwork.Interface, err error) {
	resourceName := "network interface"
	nsg, err := GetNetworkSecurityGroup(cntx, nsgName, tenantID, clientID, clientSecret, subscriptionID, rgName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get netwrok security group: %v", err)))
	}
	subnet, err := GetVirtualNetworkSubnet(cntx, vnetName, subnetName, tenantID, clientID, clientSecret, subscriptionID, rgName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get subnet: %v", err)))
	}
	ip, err := GetPublicIP(cntx, ipName, tenantID, clientID, clientSecret, subscriptionID, rgName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get ip address: %v", err)))
	}
	nicClient, err := getNicClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return nic, err
	}
	future, err := nicClient.BeginCreateOrUpdate(
		cntx,
		rgName,
		netInterfaceName,
		armnetwork.Interface{
			Name:     to.Ptr(netInterfaceName),
			Location: to.Ptr(location),
			Properties: &armnetwork.InterfacePropertiesFormat{
				NetworkSecurityGroup: &nsg,
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.Ptr("ipConfig1"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							Subnet:                    &subnet,
							PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
							PublicIPAddress:           &ip,
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	resp, err := future.PollUntilDone(cntx, nil)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get nic create or update future response: %v", err)))
	}
	return resp.Interface, nil
}

func GetNetworkSecurityGroup(cntx context.Context, nsgName, tenantID, clientID, clientSecret, subscriptionID, rgName string) (armnetwork.SecurityGroup, error) {
	nsgClient, err := getNsgClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return armnetwork.SecurityGroup{}, err
	}
	resp, err := nsgClient.Get(cntx, rgName, nsgName, nil)
	if err != nil {
		return armnetwork.SecurityGroup{}, err
	}
	return resp.SecurityGroup, nil
}

func GetVirtualNetworkSubnet(cntx context.Context, vnetName string, subnetName, tenantID, clientID, clientSecret, subscriptionID, rgName string) (armnetwork.Subnet, error) {
	subnetsClient, err := getSubnetsClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return armnetwork.Subnet{}, err
	}
	resp, err := subnetsClient.Get(cntx, rgName, vnetName, subnetName, nil)
	if err != nil {
		return armnetwork.Subnet{}, err
	}
	return resp.Subnet, nil
}

func GetPublicIP(cntx context.Context, ipName, tenantID, clientID, clientSecret, subscriptionID, rgName string) (armnetwork.PublicIPAddress, error) {
	ipClient, err := getIPClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return armnetwork.PublicIPAddress{}, err
	}
	resp, err := ipClient.Get(cntx, rgName, ipName, nil)
	if err != nil {
		return armnetwork.PublicIPAddress{}, err
	}
	return resp.PublicIPAddress, nil
}

func GetNic(cntx context.Context, nicName, tenantID, clientID, clientSecret, subscriptionID, rgName string) (armnetwork.Interface, error) {
	nicClient, err := getNicClient(tenantID, clientID, clientSecret, subscriptionID)
	if err != nil {
		return armnetwork.Interface{}, err
	}
	resp, err := nicClient.Get(cntx, rgName, nicName, nil)
	if err != nil {
		return armnetwork.Interface{}, err
	}
	return resp.Interface, nil
}
