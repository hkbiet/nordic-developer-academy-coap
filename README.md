# Nordic Developer Academy: CoAP test server

Documents how to set up the CoAP test server and which features it must provide.

## Persistence

The dynamic resources are persisted in Azure Blob Storage. Up to 100 KB are allowed. Uploads are deleted after 24 hours using [lifecycle management](https://learn.microsoft.com/en-us/azure/storage/blobs/lifecycle-management-overview).

Configure these environment variables:

- `STORAGE_CONNECTION_STRING`: the connection string to use, e.g. `DefaultEndpointsProtocol=https;AccountName=myAccountName;AccountKey=myAccountKey;EndpointSuffix=core.windows.net`
- `STORAGE_CONTAINER_NAME`: the container name to use, e.g. `golang`
