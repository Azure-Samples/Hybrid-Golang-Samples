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
- List virtual machines
- Delete a virtual machine
- Create a managed disk 
- Create a managed disk VM

To see the code to perform these operations,
check out the `main()` function in [app.go](app.go).
Each operation is clearly labeled with a comment and a print function.

## Running this sample

1. Open a Powershell or Bash shell in `...\Hybrid-Golang-Samples\keyvault` and enter the following commands:

    ```powershell
    go mod tidy
    ```

1. Run the sample.

    ```powershell
    go run app.go [-secret] [-clean] [-disableID]
    ```

    -clean deletes the resource group created during the run

    -secret uses the secret config file

    -disableID disables instance discovery

## More information

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](http://go.microsoft.com/fwlink/?LinkId=330212).

---

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

