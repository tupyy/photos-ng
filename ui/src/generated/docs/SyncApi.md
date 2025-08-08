# SyncApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getSyncJob**](#getsyncjob) | **GET** /sync/{id} | Get sync job by ID|
|[**listSyncJobs**](#listsyncjobs) | **GET** /sync | List all sync jobs|
|[**startSyncJob**](#startsyncjob) | **POST** /sync | Start sync job|
|[**stopAllSyncJobs**](#stopallsyncjobs) | **DELETE** /sync | Stop all sync jobs|
|[**stopSyncJob**](#stopsyncjob) | **DELETE** /sync/{id} | Stop sync job by ID|

# **getSyncJob**
> SyncJob getSyncJob()

Retrieve detailed information about a specific sync job

### Example

```typescript
import {
    SyncApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new SyncApi(configuration);

let id: string; //The ID of the sync job to retrieve (default to undefined)

const { status, data } = await apiInstance.getSyncJob(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the sync job to retrieve | defaults to undefined|


### Return type

**SyncJob**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successful response |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listSyncJobs**
> ListSyncJobsResponse listSyncJobs()

Retrieve a list of all running and completed sync jobs

### Example

```typescript
import {
    SyncApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new SyncApi(configuration);

const { status, data } = await apiInstance.listSyncJobs();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ListSyncJobsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successful response |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **startSyncJob**
> StartSyncResponse startSyncJob(startSyncRequest)

Start an asynchronous sync job for a specified path

### Example

```typescript
import {
    SyncApi,
    Configuration,
    StartSyncRequest
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new SyncApi(configuration);

let startSyncRequest: StartSyncRequest; //

const { status, data } = await apiInstance.startSyncJob(
    startSyncRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **startSyncRequest** | **StartSyncRequest**|  | |


### Return type

**StartSyncResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**202** | Sync job started successfully |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **stopAllSyncJobs**
> StopAllSyncJobs200Response stopAllSyncJobs()

Stop all running sync jobs

### Example

```typescript
import {
    SyncApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new SyncApi(configuration);

const { status, data } = await apiInstance.stopAllSyncJobs();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**StopAllSyncJobs200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | All sync jobs stopped successfully |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **stopSyncJob**
> StopSyncJob200Response stopSyncJob()

Stop a specific sync job by its ID

### Example

```typescript
import {
    SyncApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new SyncApi(configuration);

let id: string; //The ID of the sync job to stop (default to undefined)

const { status, data } = await apiInstance.stopSyncJob(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the sync job to stop | defaults to undefined|


### Return type

**StopSyncJob200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Sync job stopped successfully |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

