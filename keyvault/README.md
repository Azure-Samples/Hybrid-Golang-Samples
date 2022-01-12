# Hybrid-Compute-GO-Create-KeyVault

These samples demonstrate interactions with key vault management in an Azure Stack environment.

To see the code to perform these operations,
check out the `main()` function in [app.go](app.go).
Each operation is clearly labeled with a comment and a print function.


## Running this sample

1. Open a Powershell or Bash shell in `...\Hybrid-Golang-Samples\keyvault` and enter the following command:
    ```
    go mod init keyvault
    ```

1. Run the following to validate the go mod file with the required source code modules.
    ```
    go mod tidy
    ```

1. Run the sample.
    ```
    go run app.go
    ```

1. Clean the resource group created during sample run.
    ```
    go run app.go clean
    ```

## More information

Here are some helpful links:

- [Azure Virtual Machines documentation](https://azure.microsoft.com/services/virtual-machines/)
- [Learning Path for Virtual Machines](https://azure.microsoft.com/documentation/learning-paths/virtual-machines/)

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](http://go.microsoft.com/fwlink/?LinkId=330212).

---

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
