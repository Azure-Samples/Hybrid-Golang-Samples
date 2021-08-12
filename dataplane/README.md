# dataplane

The code provided shows how to do the following:

- Create a resource group
- Create a storage account
- Create a container in the storage account
- Upload a file to the container

To see the code to perform these operations,
check out the `main()` function in [app.go](app.go).


## Running this sample

1.  Open a Powershell or Bash shell in `...\Hybrid-Golang-Samples\dataplane` and enter the following commands:
    ```
    go mod init dataplane
    go mod tidy
    go get github.com/Azure/azure-storage-blob-go/azblob@v0.10.0
    ```

    NOTE: The azblob@v0.10.0 version is required for AzureStack.

1. Run the sample.
    ```
    go run app.go
    ```

1. Clean the resource group created during sample run.
    ```
    go run app.go clean
    ```

## More information

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](http://go.microsoft.com/fwlink/?LinkId=330212).

---

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
