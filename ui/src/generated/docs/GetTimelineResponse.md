# GetTimelineResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**buckets** | [**Array&lt;Bucket&gt;**](Bucket.md) |  | [default to undefined]
**years** | **Array&lt;number&gt;** | List of all years that contain photos | [default to undefined]
**total** | **number** | Total number of buckets | [default to undefined]
**limit** | **number** | Number of buckets returned | [default to undefined]
**offset** | **number** | Number of buckets skipped | [default to undefined]

## Example

```typescript
import { GetTimelineResponse } from 'photos-ng-api-client';

const instance: GetTimelineResponse = {
    buckets,
    years,
    total,
    limit,
    offset,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
