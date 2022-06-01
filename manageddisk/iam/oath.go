// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package iam

import (
	"io/ioutil"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const (
	samplesAppID = "bee3737f-b06f-444f-b3c3-5b0f3fce46ea"
)

func GetResourceManagementToken(tenantID, clientID, certPass, certPath string) (azcore.TokenCredential, error) {
	certData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	certs, key, err := azidentity.ParseCertificates(certData, []byte(certPass))
	if err != nil {
		return nil, err
	}

	token, err := azidentity.NewClientCertificateCredential(tenantID, clientID, certs, key, nil)
	return token, err
}
