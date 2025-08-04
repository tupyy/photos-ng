# UpdateMediaRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**capturedAt** | **string** | Date when the media was captured | [optional] [default to undefined]
**exif** | [**Array&lt;ExifHeader&gt;**](ExifHeader.md) | EXIF data for the media | [optional] [default to undefined]

## Example

```typescript
import { UpdateMediaRequest } from 'photos-ng-api-client';

const instance: UpdateMediaRequest = {
    capturedAt,
    exif,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
