# dataplane

The code provided shows how to do the following:

- Create a Client Secret Credential
- Create a resource group
- List resource groups
- Delete a resource group

To see the code to perform these operations,
check out the `main()` function in [app.go](app.go).


## Running this sample

1.  Open a Powershell or Bash shell in `...\Hybrid-Golang-Samples\dataplane` and enter the following commands:
    ```
    go mod tidy
    ```

1. Run the sample.
    ```
    go run app.go [-secret] [-clean] [-disableID]
    ```

    -clean deletes the resource group created during the run

    -secret uses the secret config file

    -disableID disables instance discovery

## More information

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](http://go.microsoft.com/fwlink/?LinkId=330212).

---

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
