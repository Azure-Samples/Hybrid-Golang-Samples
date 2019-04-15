// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package hybridnetwork

import (
	"context"
	"fmt"
	"log"

	"../iam"

	"github.com/Azure/azure-sdk-for-go/profiles/2018-03-01/network/mgmt/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	errorPrefix = "Cannot create %v, reason: %v"
)

func getVnetClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID string) network.VirtualNetworksClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "virtual network", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	vnetClient := network.NewVirtualNetworksClientWithBaseURI(armEndpoint, subscriptionID)
	vnetClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return vnetClient
}

func getNsgClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID string) network.SecurityGroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "security group", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	nsgClient := network.NewSecurityGroupsClientWithBaseURI(armEndpoint, subscriptionID)
	nsgClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return nsgClient
}

func getIPClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID string) network.PublicIPAddressesClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "public IP address", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	ipClient := network.NewPublicIPAddressesClientWithBaseURI(armEndpoint, subscriptionID)
	ipClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return ipClient
}

func getNicClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID string) network.InterfacesClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "network interface", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	nicClient := network.NewInterfacesClientWithBaseURI(armEndpoint, subscriptionID)
	nicClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return nicClient
}

func getSubnetsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID string) network.SubnetsClient {
	token, err := iam.GetResourceManagementTokenHybrid(armEndpoint, tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "subnet", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	subnetsClient := network.NewSubnetsClientWithBaseURI(armEndpoint, subscriptionID)
	subnetsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return subnetsClient
}

// CreateVirtualNetworkAndSubnets creates a virtual network with two subnets
func CreateVirtualNetworkAndSubnets(cntx context.Context, vnetName, subnetName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName, location string) (vnet network.VirtualNetwork, err error) {
	resourceName := "virtual network and subnet"
	vnetClient := getVnetClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	future, err := vnetClient.CreateOrUpdate(
		cntx,
		rgName,
		vnetName,
		network.VirtualNetwork{
			Location: to.StringPtr(location),
			VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
				AddressSpace: &network.AddressSpace{
					AddressPrefixes: &[]string{"10.0.0.0/8"},
				},
				Subnets: &[]network.Subnet{
					{
						Name: to.StringPtr(subnetName),
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.0.0.0/16"),
						},
					},
				},
			},
		})

	if err != nil {
		return vnet, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	err = future.WaitForCompletion(cntx, vnetClient.Client)
	if err != nil {
		return vnet, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get the vnet create or update future response: %v", err)))
	}

	return future.Result(vnetClient)
}

// CreateNetworkSecurityGroup creates a new network security group
func CreateNetworkSecurityGroup(cntx context.Context, nsgName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName, location string) (nsg network.SecurityGroup, err error) {
	resourceName := "security group"
	nsgClient := getNsgClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	future, err := nsgClient.CreateOrUpdate(
		cntx,
		rgName,
		nsgName,
		network.SecurityGroup{
			Location: to.StringPtr(location),
			SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
				SecurityRules: &[]network.SecurityRule{
					{
						Name: to.StringPtr("allow_ssh"),
						SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
							Protocol:                 network.SecurityRuleProtocolTCP,
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("22"),
							Access:                   network.SecurityRuleAccessAllow,
							Direction:                network.SecurityRuleDirectionInbound,
							Priority:                 to.Int32Ptr(100),
						},
					},
					{
						Name: to.StringPtr("allow_https"),
						SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
							Protocol:                 network.SecurityRuleProtocolTCP,
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("443"),
							Access:                   network.SecurityRuleAccessAllow,
							Direction:                network.SecurityRuleDirectionInbound,
							Priority:                 to.Int32Ptr(200),
						},
					},
				},
			},
		},
	)

	if err != nil {
		return nsg, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	err = future.WaitForCompletion(cntx, nsgClient.Client)
	if err != nil {
		return nsg, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get nsg create or update future response: %v", err)))
	}

	return future.Result(nsgClient)
}

// CreatePublicIP creates a new public IP
func CreatePublicIP(cntx context.Context, ipName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName, location string) (ip network.PublicIPAddress, err error) {
	resourceName := "public IP"
	ipClient := getIPClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	future, err := ipClient.CreateOrUpdate(
		cntx,
		rgName,
		ipName,
		network.PublicIPAddress{
			Name:     to.StringPtr(ipName),
			Location: to.StringPtr(location),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: network.Static,
			},
		},
	)

	if err != nil {
		return ip, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	err = future.WaitForCompletion(cntx, ipClient.Client)
	if err != nil {
		return ip, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get public ip address create or update future response: %v", err)))
	}
	return future.Result(ipClient)
}

// CreateNetworkInterface creates a new network interface
func CreateNetworkInterface(cntx context.Context, netInterfaceName, nsgName, vnetName, subnetName, ipName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName, location string) (nic network.Interface, err error) {
	resourceName := "network interface"
	nsg, err := GetNetworkSecurityGroup(cntx, nsgName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get netwrok security group: %v", err)))
	}
	subnet, err := GetVirtualNetworkSubnet(cntx, vnetName, subnetName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get subnet: %v", err)))
	}
	ip, err := GetPublicIP(cntx, ipName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get ip address: %v", err)))
	}
	nicClient := getNicClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	future, err := nicClient.CreateOrUpdate(
		cntx,
		rgName,
		netInterfaceName,
		network.Interface{
			Name:     to.StringPtr(netInterfaceName),
			Location: to.StringPtr(location),
			InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
				NetworkSecurityGroup: &nsg,
				IPConfigurations: &[]network.InterfaceIPConfiguration{
					{
						Name: to.StringPtr("ipConfig1"),
						InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
							Subnet: &subnet,
							PrivateIPAllocationMethod: network.Dynamic,
							PublicIPAddress:           &ip,
						},
					},
				},
			},
		},
	)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}
	err = future.WaitForCompletion(cntx, nicClient.Client)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get nic create or update future response: %v", err)))
	}
	return future.Result(nicClient)
}

func GetNetworkSecurityGroup(cntx context.Context, nsgName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName string) (network.SecurityGroup, error) {
	nsgClient := getNsgClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	return nsgClient.Get(cntx, rgName, nsgName, "")
}

func GetVirtualNetworkSubnet(cntx context.Context, vnetName string, subnetName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName string) (network.Subnet, error) {
	subnetsClient := getSubnetsClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	return subnetsClient.Get(cntx, rgName, vnetName, subnetName, "")
}

func GetPublicIP(cntx context.Context, ipName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName string) (network.PublicIPAddress, error) {
	ipClient := getIPClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	return ipClient.Get(cntx, rgName, ipName, "")
}

func GetNic(cntx context.Context, nicName, tenantID, clientID, clientSecret, armEndpoint, subscriptionID, rgName string) (network.Interface, error) {
	nicClient := getNicClient(tenantID, clientID, clientSecret, armEndpoint, subscriptionID)
	return nicClient.Get(cntx, rgName, nicName, "")
}
