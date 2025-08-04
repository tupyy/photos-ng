# TimelineApi

All URIs are relative to *http://localhost:8080/api/v1*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getTimeline**](#gettimeline) | **GET** /timeline | Get timeline buckets|

# **getTimeline**
> GetTimelineResponse getTimeline()

Retrieve a timeline of media organized in buckets

### Example

```typescript
import {
    TimelineApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new TimelineApi(configuration);

let startDate: string; //Start date for the timeline (optional) (default to undefined)
let endDate: string; //Start date for the timeline (optional) (default to undefined)
let limit: number; //Maximum number of buckets to return (optional) (default to 20)
let offset: number; //Number of buckets to skip (optional) (default to 0)

const { status, data } = await apiInstance.getTimeline(
    startDate,
    endDate,
    limit,
    offset
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **startDate** | [**string**] | Start date for the timeline | (optional) defaults to undefined|
| **endDate** | [**string**] | Start date for the timeline | (optional) defaults to undefined|
| **limit** | [**number**] | Maximum number of buckets to return | (optional) defaults to 20|
| **offset** | [**number**] | Number of buckets to skip | (optional) defaults to 0|


### Return type

**GetTimelineResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successful response |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

