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

1. If you don't already have it, [install Golang](https://golang.org/doc/install).

1. Install Go SDK and its dependencies, [install Go SDK](https://github.com/azure/azure-sdk-for-go) 

1. Clone the sample project repository to your `GOPATH` location. You can add a new path to `GOPATH` location by adding an existing folder path to the `GOPATH` user environment variable. 
    - Create a `src` folder inside this new `GOPATH` folder and `cd` into the `src` folder.
    ```
    mkdir src
    cd src
    ```
    - Clone the sample project repository into your `src` folder.
    ```
    git clone https://github.com/Azure-Samples/Hybrid-Compute-Go-ManagedDisks.git
    ```

1. Open a Powershell or Bash shell in `$GOPATH/src/Hybrid-Compute-Go-Create-VM` and enter the following command:
    ```
    go mod init Hybrid-Compute-Go-Create-VM
    ```

1. Run the following to validate the go mod file with the required source code modules.
    ```
    go mod tidy
    ```

1. Create a [service principal](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-create-service-principals) to work against AzureStack. Make sure your service principal has [contributor/owner role](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-create-service-principals#assign-role-to-service-principal) on your subscription.

1. Fill in and export these environment variables into your current shell. 
    ```
    export AZURE_ARM_ENDPOINT={your AzureStack Resource Manager Endpoint}
    export AZURE_TENANT_ID={your tenant id for AAD or "adfs" for ADFS}
    export AZURE_SP_APP_ID={your service principal client id that came with your service principal client secret}
    export AZURE_SP_APP_SECRET ={your service principal client secret}
    export AZURE_SUBSCRIPTION_ID={your subscription id}
    export AZURE_LOCATION={your resource location}
    ```

1. Note that in order to run this sample, Canonial UbuntuServer 16.04-LTS image must be present in AzureStack market place. These can be [downloaded from Azure](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-download-azure-marketplace-item) and [added to Platform Image Repository](https://docs.microsoft.com/en-us/azure/azure-stack/azure-stack-add-vm-image).

1. Azure will force the use of SSH over password authentication if both were configured. The sample code also enforces SSH key pair authentication over password authentication. The password authentication will be used if the SSH key pair path does not exist. To run the sample, do one of either:
    - Create an SSH key at `%HOMEPATH%/.ssh/id_rsa.pub` for SSH key pair authentication.
    ```
    go run app.go
    ```
    - Pass a string parameter as the VM admin password for password authentication.
    ```
    go run app.go <PASSWORD>
    ```


## More information

Here are some helpful links:

- [Azure Virtual Machines documentation](https://azure.microsoft.com/services/virtual-machines/)
- [Learning Path for Virtual Machines](https://azure.microsoft.com/documentation/learning-paths/virtual-machines/)

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](http://go.microsoft.com/fwlink/?LinkId=330212).

---

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
