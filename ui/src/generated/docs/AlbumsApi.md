# AlbumsApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createAlbum**](#createalbum) | **POST** /albums | Create a new album|
|[**deleteAlbum**](#deletealbum) | **DELETE** /albums/{id} | Delete album by ID|
|[**getAlbum**](#getalbum) | **GET** /albums/{id} | Get album by ID|
|[**listAlbums**](#listalbums) | **GET** /albums | List all albums|
|[**syncAlbum**](#syncalbum) | **POST** /albums/{id}/sync | Sync album|
|[**updateAlbum**](#updatealbum) | **PUT** /albums/{id} | Update album by ID|

# **createAlbum**
> Album createAlbum(createAlbumRequest)

Create a new album

### Example

```typescript
import {
    AlbumsApi,
    Configuration,
    CreateAlbumRequest
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new AlbumsApi(configuration);

let createAlbumRequest: CreateAlbumRequest; //

const { status, data } = await apiInstance.createAlbum(
    createAlbumRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createAlbumRequest** | **CreateAlbumRequest**|  | |


### Return type

**Album**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Album created successfully |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteAlbum**
> deleteAlbum()

Delete a specific album by its ID

### Example

```typescript
import {
    AlbumsApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new AlbumsApi(configuration);

let id: string; //The ID of the album to delete (default to undefined)

const { status, data } = await apiInstance.deleteAlbum(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the album to delete | defaults to undefined|


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | Album deleted successfully |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAlbum**
> Album getAlbum()

Retrieve a specific album by its ID

### Example

```typescript
import {
    AlbumsApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new AlbumsApi(configuration);

let id: string; //The ID of the album to retrieve (default to undefined)

const { status, data } = await apiInstance.getAlbum(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the album to retrieve | defaults to undefined|


### Return type

**Album**

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

# **listAlbums**
> ListAlbumsResponse listAlbums()

Retrieve a list of all albums

### Example

```typescript
import {
    AlbumsApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new AlbumsApi(configuration);

let limit: number; //Maximum number of albums to return (optional) (default to 20)
let offset: number; //Number of albums to skip (optional) (default to 0)

const { status, data } = await apiInstance.listAlbums(
    limit,
    offset
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **limit** | [**number**] | Maximum number of albums to return | (optional) defaults to 20|
| **offset** | [**number**] | Number of albums to skip | (optional) defaults to 0|


### Return type

**ListAlbumsResponse**

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

# **syncAlbum**
> SyncAlbumResponse syncAlbum()

Synchronize an album with the file system

### Example

```typescript
import {
    AlbumsApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new AlbumsApi(configuration);

let id: string; //The ID of the album to sync (default to undefined)

const { status, data } = await apiInstance.syncAlbum(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the album to sync | defaults to undefined|


### Return type

**SyncAlbumResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Album sync completed successfully |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateAlbum**
> Album updateAlbum(updateAlbumRequest)

Update a specific album by its ID

### Example

```typescript
import {
    AlbumsApi,
    Configuration,
    UpdateAlbumRequest
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new AlbumsApi(configuration);

let id: string; //The ID of the album to update (default to undefined)
let updateAlbumRequest: UpdateAlbumRequest; //

const { status, data } = await apiInstance.updateAlbum(
    id,
    updateAlbumRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateAlbumRequest** | **UpdateAlbumRequest**|  | |
| **id** | [**string**] | The ID of the album to update | defaults to undefined|


### Return type

**Album**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Album updated successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

