---
page_type: sample
languages:
- go
products:
- azure
description: "These samples demonstrate how to create Virtual Machines using the Azure SDK for Go on Azure Stack."
urlFragment: Hybrid-Compute-Go-Create-VM
---

# Hybrid-Compute-GO-Create-VM

These samples demonstrate how to create Virtual Machines using the Azure SDK for Go on Azure Stack.
The code provided shows how to do the following:

- Create a resource group
- Create a virtual network
- Create a security group
- Create a public IP
- Create a network interface
- Create a storage account
- Create a virtual machine

To see the code to perform these operations,
check out the `main()` function in [app.go](app.go).
Each operation is clearly labeled with a comment and a print function.


## Running this sample
1.  If you don't already have it, [install Golang](https://golang.org/doc/install).

2.  Install Go SDK and its dependencies, [install Go SDK](https://github.com/azure/azure-sdk-for-go) 

3.  Clone the repository.

    ```
    git clone https://github.com/Azure-Samples/Hybrid-Compute-Go-Create-VM.git
    ```

4.  Move the Hybrid-Compute-Go-Create-VM folder to your $GOPATH/src folder.

5.  Open a Powershell or Bash shell in $GOPATH/src/Hybrid-Compute-Go-Create-VM and enter the following command:

    ```
    go mod init Hybrid-Compute-Go-Create-VM
    ```

6.  Create a [service principal](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-create-service-principals) to work against AzureStack. Make sure your service principal has [contributor/owner role](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-create-service-principals#assign-role-to-service-principal) on your subscription.

7.  Fill in and export these environment variables into your current shell. 

    ```
    export AZS_ARM_ENDPOINT={your AzureStack Resource Manager Endpoint}
    export AZS_TENANT_ID={your tenant id}
    export AZS_SECRET_CLIENT_ID={your service principal client id that came with your service principal client secret}
    export AZS_CLIENT_SECRET={your service principal client secret}
    export AZS_SUBSCRIPTION_ID={your subscription id}
    export AZS_LOCATION={your resource location}
    
    ```

8.  Note that in order to run this sample, Canonial UbuntuServer 16.04-LTS image must be present in AzureStack market place. These can be either [downloaded from Azure](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-download-azure-marketplace-item) or [added to Platform Image Repository](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-add-vm-image).


9.  Run the sample.

    ```
    go run app.go
    ```
    
10. To clean up resources, sign into the Service Principal used to create the resource group "stackrg" (assigned in app.go) and remove the resource group:

    ```
    $clientIdForSecret = "<Client Id associated with a client secret>"
    $clientSecret = "<Client secret>"
    $tenantId = "<The tenant Id>"
    $servicePrincipalSecurePassword = $clientSecret | ConvertTo-SecureString -AsPlaintText -Force
    $servicePrincipalCredential = New-Object -TypeName System.Management.Automation.PSCredential `
        -ArgumentList $clientIdForSecret $servicePrincipalSecurePassword
    Connect-AzAccount -Environment $ServiceAdminEnvironmentName `
        -ServicePrincipal `
        -Credential $servicePrincipalCredential `
        -TenantId $tenantID
    Remove-AzResourceGroup -Name "stackrg"
    ```

## More information

Here are some helpful links:

- [Azure Virtual Machines documentation](https://azure.microsoft.com/services/virtual-machines/)
- [Learning Path for Virtual Machines](https://azure.microsoft.com/documentation/learning-paths/virtual-machines/)

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](http://go.microsoft.com/fwlink/?LinkId=330212).

---

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
