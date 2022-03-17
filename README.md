---
page_type: sample
languages: 
- go
products: 
- azure-sdks
description: "These samples demonstrate various interaction with Azure Stack Hub."
urlFragment: Hybrid-Golang-Samples
---

# Hybrid-Golang-Samples

<!-- 
Guidelines on README format: https://review.docs.microsoft.com/help/onboard/admin/samples/concepts/readme-template?branch=master

Guidance on onboarding samples to docs.microsoft.com/samples: https://review.docs.microsoft.com/help/onboard/admin/samples/process/onboarding?branch=master

Taxonomies for products and languages: https://review.docs.microsoft.com/new-hope/information-architecture/metadata/taxonomies?branch=master
-->

This repository is for Azure Stack Hub JavaScript samples. Each of the sub-directories contain README.md files detailing how to that sample.

If you don't already have it, [install Golang](https://golang.org/doc/install).

## Create Service Principal
Create a [service principal](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-create-service-principals) to work against AzureStack. Make sure your service principal has [contributor/owner role](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-create-service-principals#assign-role-to-service-principal) on your subscription. The samples use either a secret or certificate service principal.

## Configure Service Principal Details
Some of the configuration parameters from service principal objects may not be used in the samples. The configuration file includes them anyway for thoroughness and future-proofing.

### Setup Secret Service Principal
1. Make a copy of `azureSecretSpConfig.json.dist` and `azureCertSpConfig.json.dist`, then rename those copies to `azureSecretSpConfig.json` and `azureCertSpConfig.json`. Each sample will use one of these configuration files.
1. Fill in the following values in the corresponding JSON files:

Set the following JSON properties in `./azureSecretSpConfig.json`.

| Variable              | Description                                                 |
|-----------------------|-------------------------------------------------------------|
| `clientId`            | Service principal application id.                            |
| `clientSecret`        | Service principal application secret.                        |
| `clientObjectId`      | Service principal object id.                                 |
| `tenantId`            | Azure Stack Hub tenant id.                                   |
| `subscriptionId`      | Subscription id used to access offers in Azure Stack Hub.    |
| `resourceManagerUrl`  | Azure Stack Hub Resource Manager Endpoint.                   |
| `location`            | Azure Resource location.                                     |

### Setup Certificate Service Principal 

The certificate service principal will be similar in output to secret service principal, except it uses `./azureCertSpConfig.json` config file.

| Variable              | Description                                                 |
|-----------------------|-------------------------------------------------------------|
| `clientId`            | Service principal application id.                            |
| `certPass`            | Certificate password                                        |
| `certPath`            | "/" separated path to the certificate.                      |
| `clientObjectId`      | Service principal object id.                                 |
| `tenantId`            | Azure Stack Hub tenant id.                                   |
| `subscriptionId`      | Subscription id used to access offers in Azure Stack Hub.    |
| `resourceManagerUrl`  | Azure Stack Hub Resource Manager Endpoint.                   |
| `location`            | Azure Resource location.                                     |

Service principal PowerShell object output example for secret service principal:

AAD
```
Secret                : System.Security.SecureString                                 # clientSecret  (decrypt for external use)
ServicePrincipalNames : {bd6bb75f-5fd6-4db9-91b7-4a6941e7feb9, http://azs-sptest01}
ApplicationId         : bd6bb75f-5fd6-4db9-91b7-4a6941e7feb9                         # clientId
DisplayName           : azs-sptest01
Id                    : 36a22ee4-e2b0-411d-8f21-0ea8b4b5c46f                         # clientObjectId
AdfsId                : 
Type                  : ServicePrincipal
```

ADFS
```
ApplicationIdentifier : S-1-5-21-2937821301-3551617933-4294865508-76632              # clientObjectId
ClientId              : 7591924e-0341-4812-8d23-52ef0aa27eff                         # clientId
Thumbprint            : 
ApplicationName       : Azurestack-azs-sptest01
ClientSecret          : <Redacted>                                                   # clientSecret
PSComputerName        : <Redacted>
RunspaceId            : e841cbbc-3d8e-45fd-b63f-42adbfbf664b
```

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
