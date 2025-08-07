# StatsApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getStats**](#getstats) | **GET** /stats | Get application statistics|

# **getStats**
> StatsResponse getStats()

Retrieve statistics including album count, media count, and years with media

### Example

```typescript
import {
    StatsApi,
    Configuration
} from 'photos-ng-api-client';

const configuration = new Configuration();
const apiInstance = new StatsApi(configuration);

const { status, data } = await apiInstance.getStats();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**StatsResponse**

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

