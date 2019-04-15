---
services: Azure-Stack
platforms: GO
author: seyadava
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

4.  Create a [service principal](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-create-service-principals) to work against AzureStack. Make sure your service principal has [contributor/owner role](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-create-service-principals#assign-role-to-service-principal) on your subscription.

6.  Fill in and export these environment variables into your current shell. 

    ```
    export AZ_ARM_ENDPOINT={your AzureStack Resource Manager Endpoint}
    export AZ_TENANT_ID={your tenant id}
    export AZ_CLIENT_ID={your client id}
    export AZ_CLIENT_SECRET={your client secret}
    export AZ_SUBSCRIPTION_ID={your subscription id}
    export AZ_LOCATION={your resource location}
    
    ```

7.  Note that in order to run this sample, WindowsServer 2012-R2-Datacenter image must be present in AzureStack market place. These can be either [downloaded from Azure](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-download-azure-marketplace-item) or [added to Platform Image Repository](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-add-vm-image).


8. Run the sample.

    ```
    go run app.go
    ```
    
## More information

Here are some helpful links:

- [Azure Virtual Machines documentation](https://azure.microsoft.com/services/virtual-machines/)
- [Learning Path for Virtual Machines](https://azure.microsoft.com/documentation/learning-paths/virtual-machines/)

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](http://go.microsoft.com/fwlink/?LinkId=330212).

---

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
