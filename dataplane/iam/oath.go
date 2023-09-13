// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package iam

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"golang.org/x/crypto/pkcs12"
)

const (
	samplesAppID = "bee3737f-b06f-444f-b3c3-5b0f3fce46ea"
)

func decodePkcs12(pkcs []byte, password string) (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, certificate, err := pkcs12.Decode(pkcs, password)
	if err != nil {
		return nil, nil, err
	}

	rsaPrivateKey, isRsaKey := privateKey.(*rsa.PrivateKey)
	if !isRsaKey {
		return nil, nil, fmt.Errorf("PKCS#12 certificate must contain an RSA private key")
	}

	return certificate, rsaPrivateKey, nil
}

// GetResourceManagementToken retrieves auth token
func GetResourceManagementToken(tenantID, clientID, certPass, armEndpoint, certPath string) (adal.OAuthTokenProvider, error) {
	var token adal.OAuthTokenProvider
	environment, err := azure.EnvironmentFromURL(armEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Cannot retrieve environment metadata properties. Error details: %v", err)
	}
	tokenAudience := environment.TokenAudience
	activeDirectoryEndpoint := environment.ActiveDirectoryEndpoint
	oauthConfig, err := adal.NewOAuthConfig(activeDirectoryEndpoint, tenantID)
	certData, err := ioutil.ReadFile(certPath)
	certificate, rsaPrivateKey, err := decodePkcs12(certData, certPass)
	if err != nil {
		return nil, err
	}
	token, err = adal.NewServicePrincipalTokenFromCertificate(
		*oauthConfig,
		clientID,
		certificate,
		rsaPrivateKey,
		tokenAudience)

	return token, err
}
