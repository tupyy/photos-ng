# ListMediaResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**media** | [**Array&lt;Media&gt;**](Media.md) |  | [default to undefined]
**limit** | **number** | Number of media items returned | [default to undefined]
**nextCursor** | **string** | Cursor for next page (base64 encoded) | [optional] [default to undefined]

## Example

```typescript
import { ListMediaResponse } from 'photos-ng-api-client';

const instance: ListMediaResponse = {
    media,
    limit,
    nextCursor,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
