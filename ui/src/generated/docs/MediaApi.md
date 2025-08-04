# MediaApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**deleteMedia**](#deletemedia) | **DELETE** /media/{id} | Delete media by ID|
|[**getMedia**](#getmedia) | **GET** /media/{id} | Get media by ID|
|[**getMediaContent**](#getmediacontent) | **GET** /media/{id}/content | Get media content|
|[**getMediaThumbnail**](#getmediathumbnail) | **GET** /media/{id}/thumbnail | Get media thumbnail|
|[**listMedia**](#listmedia) | **GET** /media | List all media|
|[**updateMedia**](#updatemedia) | **PUT** /media/{id} | Update media by ID|

# **deleteMedia**
> deleteMedia()

Delete a specific media item by its ID

### Example

```typescript
import {
    MediaApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let id: string; //The ID of the media item to delete (default to undefined)

const { status, data } = await apiInstance.deleteMedia(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the media item to delete | defaults to undefined|


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
|**204** | Media deleted successfully |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMedia**
> Media getMedia()

Retrieve a specific media item by its ID

### Example

```typescript
import {
    MediaApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let id: string; //The ID of the media item to retrieve (default to undefined)

const { status, data } = await apiInstance.getMedia(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the media item to retrieve | defaults to undefined|


### Return type

**Media**

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

# **getMediaContent**
> File getMediaContent()

Retrieve the full content of a media item

### Example

```typescript
import {
    MediaApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let id: string; //The ID of the media item (default to undefined)

const { status, data } = await apiInstance.getMediaContent(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the media item | defaults to undefined|


### Return type

**File**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: image/*, video/*, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successful response |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMediaThumbnail**
> File getMediaThumbnail()

Retrieve the thumbnail image for a media item

### Example

```typescript
import {
    MediaApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let id: string; //The ID of the media item (default to undefined)

const { status, data } = await apiInstance.getMediaThumbnail(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the media item | defaults to undefined|


### Return type

**File**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: image/*, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successful response |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listMedia**
> ListMediaResponse listMedia()

Retrieve a list of all media items

### Example

```typescript
import {
    MediaApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let limit: number; //Maximum number of media items to return (optional) (default to 20)
let offset: number; //Number of media items to skip (optional) (default to 0)
let albumId: string; //Filter media by album ID (optional) (default to undefined)
let type: 'photo' | 'video'; //Filter media by type (optional) (default to undefined)
let startDate: string; //Filter media captured on or after this date (optional) (default to undefined)
let endDate: string; //Filter media captured on or before this date (optional) (default to undefined)
let sortBy: 'capturedAt' | 'filename' | 'type'; //Sort media by field (optional) (default to 'capturedAt')
let sortOrder: 'asc' | 'desc'; //Sort order (optional) (default to 'desc')

const { status, data } = await apiInstance.listMedia(
    limit,
    offset,
    albumId,
    type,
    startDate,
    endDate,
    sortBy,
    sortOrder
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **limit** | [**number**] | Maximum number of media items to return | (optional) defaults to 20|
| **offset** | [**number**] | Number of media items to skip | (optional) defaults to 0|
| **albumId** | [**string**] | Filter media by album ID | (optional) defaults to undefined|
| **type** | [**&#39;photo&#39; | &#39;video&#39;**]**Array<&#39;photo&#39; &#124; &#39;video&#39;>** | Filter media by type | (optional) defaults to undefined|
| **startDate** | [**string**] | Filter media captured on or after this date | (optional) defaults to undefined|
| **endDate** | [**string**] | Filter media captured on or before this date | (optional) defaults to undefined|
| **sortBy** | [**&#39;capturedAt&#39; | &#39;filename&#39; | &#39;type&#39;**]**Array<&#39;capturedAt&#39; &#124; &#39;filename&#39; &#124; &#39;type&#39;>** | Sort media by field | (optional) defaults to 'capturedAt'|
| **sortOrder** | [**&#39;asc&#39; | &#39;desc&#39;**]**Array<&#39;asc&#39; &#124; &#39;desc&#39;>** | Sort order | (optional) defaults to 'desc'|


### Return type

**ListMediaResponse**

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

# **updateMedia**
> Media updateMedia(updateMediaRequest)

Update a specific media item by its ID

### Example

```typescript
import {
    MediaApi,
    Configuration,
    UpdateMediaRequest
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new MediaApi(configuration);

let id: string; //The ID of the media item to update (default to undefined)
let updateMediaRequest: UpdateMediaRequest; //

const { status, data } = await apiInstance.updateMedia(
    id,
    updateMediaRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateMediaRequest** | **UpdateMediaRequest**|  | |
| **id** | [**string**] | The ID of the media item to update | defaults to undefined|


### Return type

**Media**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Media updated successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

