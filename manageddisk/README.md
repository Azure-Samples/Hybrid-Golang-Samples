# manageddisk

These samples demonstrate how to create Virtual Machines with managed disks using the Azure SDK for Go on Azure Stack.
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

1. Open a Powershell or Bash shell in `...\Hybrid-Golang-Samples\manageddisks` and enter the following command:
    ```
    go mod init manageddisk
    ```

1. Run the following to validate the go mod file with the required source code modules.
    ```
    go mod tidy
    ```

1. Azure will force the use of SSH over password authentication if both were configured. The sample code also enforces SSH key pair authentication over password authentication. The password authentication will be used if the SSH key pair path does not exist. To run the sample, do one of either:
    - Create an SSH key at `%HOMEPATH%/.ssh/id_rsa.pub` for SSH key pair authentication.
    ```
    go run app.go
    ```
    - Pass a string parameter as the VM admin password for password authentication.
    ```
    go run app.go <PASSWORD>
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
