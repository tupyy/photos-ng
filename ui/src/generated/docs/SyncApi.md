# SyncApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**actionAllSyncJobs**](#actionallsyncjobs) | **PATCH** /sync | Perform action on all sync jobs|
|[**actionSyncJob**](#actionsyncjob) | **PATCH** /sync/{id} | Perform action on sync job by ID|
|[**clearFinishedSyncJobs**](#clearfinishedsyncjobs) | **DELETE** /sync | Clear finished sync jobs|
|[**getSyncJob**](#getsyncjob) | **GET** /sync/{id} | Get sync job by ID|
|[**listSyncJobs**](#listsyncjobs) | **GET** /sync | List all sync jobs|
|[**startSyncJob**](#startsyncjob) | **POST** /sync | Start sync job|
|[**stopSyncJob**](#stopsyncjob) | **DELETE** /sync/{id} | Stop sync job by ID (deprecated)|

# **actionAllSyncJobs**
> SyncJobActionResponse actionAllSyncJobs(syncJobActionRequest)

Perform an action (stop or resume) on all applicable sync jobs

### Example

```typescript
import {
    SyncApi,
    Configuration,
    SyncJobActionRequest
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new SyncApi(configuration);

let syncJobActionRequest: SyncJobActionRequest; //

const { status, data } = await apiInstance.actionAllSyncJobs(
    syncJobActionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **syncJobActionRequest** | **SyncJobActionRequest**|  | |


### Return type

**SyncJobActionResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Action performed successfully on sync jobs |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **actionSyncJob**
> SyncJobActionResponse actionSyncJob(syncJobActionRequest)

Perform an action (stop or resume) on a specific sync job

### Example

```typescript
import {
    SyncApi,
    Configuration,
    SyncJobActionRequest
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new SyncApi(configuration);

let id: string; //The ID of the sync job to perform action on (default to undefined)
let syncJobActionRequest: SyncJobActionRequest; //

const { status, data } = await apiInstance.actionSyncJob(
    id,
    syncJobActionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **syncJobActionRequest** | **SyncJobActionRequest**|  | |
| **id** | [**string**] | The ID of the sync job to perform action on | defaults to undefined|


### Return type

**SyncJobActionResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Action performed successfully on sync job |  -  |
|**400** | Bad request |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **clearFinishedSyncJobs**
> ClearFinishedSyncJobsResponse clearFinishedSyncJobs()

Remove all completed, stopped, and failed sync jobs from the system

### Example

```typescript
import {
    SyncApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new SyncApi(configuration);

const { status, data } = await apiInstance.clearFinishedSyncJobs();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ClearFinishedSyncJobsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Finished sync jobs cleared successfully |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

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

