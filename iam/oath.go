// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package iam

import (
	"github.com/Azure/go-autorest/autorest/adal"
)

const (
	samplesAppID = "bee3737f-b06f-444f-b3c3-5b0f3fce46ea"
)

// OAuthGrantType specifies which grant type to use.
type OAuthGrantType int

const (
	// OAuthGrantTypeServicePrincipal for client credentials flow
	OAuthGrantTypeServicePrincipal OAuthGrantType = iota
	// OAuthGrantTypeDeviceFlow for device-auth flow
	OAuthGrantTypeDeviceFlow
)

// GetResourceManagementTokenHybrid retrieves auth token for hybrid environment
func GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tenantID, clientID, clientSecret, activeDirectoryResourceID string) (adal.OAuthTokenProvider, error) {
	var token adal.OAuthTokenProvider
	oauthConfig, err := adal.NewOAuthConfig(activeDirectoryEndpoint, tenantID)
	token, err = adal.NewServicePrincipalToken(
		*oauthConfig,
		clientID,
		clientSecret,
		activeDirectoryResourceID)

	return token, err
}
