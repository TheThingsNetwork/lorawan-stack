// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blob

// This code has been copy-pasted from https://github.com/google/go-cloud/blob/master/blob/azureblob/azureblob.go
// because the gocloud.dev package is pinned to version v0.19.0, which doesn't support MSI.

import (
	"strconv"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"gocloud.dev/blob/azureblob"
)

const (
	tokenRefreshTolerance = 300
)

// openerFromMSI acquires an MSI token and returns TokenCredential backed URLOpener
func openerFromMSI(accountName azureblob.AccountName, clientID ClientID, opts azureblob.Options) (*azureblob.URLOpener, error) {
	spToken, err := getMSIServicePrincipalToken(azure.PublicCloud.ResourceIdentifiers.Storage, clientID)
	if err != nil {
		return nil, err
	}

	err = spToken.Refresh()
	if err != nil {
		return nil, err
	}

	credential := azblob.NewTokenCredential(spToken.Token().AccessToken, defaultTokenRefreshFunction(spToken))
	return &azureblob.URLOpener{
		AccountName: accountName,
		Pipeline:    azureblob.NewPipeline(credential, azblob.PipelineOptions{}),
		Options:     opts,
	}, nil
}

var defaultTokenRefreshFunction = func(spToken *adal.ServicePrincipalToken) func(credential azblob.TokenCredential) time.Duration {
	return func(credential azblob.TokenCredential) time.Duration {
		err := spToken.Refresh()
		if err != nil {
			return 0
		}
		expiresIn, err := strconv.ParseInt(string(spToken.Token().ExpiresIn), 10, 64)
		if err != nil {
			return 0
		}
		credential.SetToken(spToken.Token().AccessToken)
		return time.Duration(expiresIn-tokenRefreshTolerance) * time.Second
	}
}

// ClientID represents the client ID, which is a specifier of particular identity to use when many are available.
type ClientID string

// getMSIServicePrincipalToken retrieves Azure API Service Principal token.
func getMSIServicePrincipalToken(resource string, clientID ClientID) (*adal.ServicePrincipalToken, error) {
	msiEndpoint, err := adal.GetMSIEndpoint()
	if err != nil {
		return nil, err
	}

	var token *adal.ServicePrincipalToken
	if clientID == "" {
		token, err = adal.NewServicePrincipalTokenFromMSI(msiEndpoint, resource)
	} else {
		opts := &adal.ManagedIdentityOptions{
			ClientID: string(clientID),
		}
		token, err = adal.NewServicePrincipalTokenFromManagedIdentity(resource, opts)
	}

	if err != nil {
		return nil, err
	}
	return token, nil
}
