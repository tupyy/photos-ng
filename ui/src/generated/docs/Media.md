# Media


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [default to undefined]
**href** | **string** |  | [default to undefined]
**albumHref** | **string** |  | [default to undefined]
**capturedAt** | **string** |  | [default to undefined]
**type** | **string** |  | [default to undefined]
**filename** | **string** | full path of the media file on the disk | [default to undefined]
**thumbnail** | **string** | href to thumbnail | [default to undefined]
**content** | **string** | href of the endpoint serving the content of the media | [default to undefined]
**exif** | [**Array&lt;ExifHeader&gt;**](ExifHeader.md) |  | [default to undefined]

## Example

```typescript
import { Media } from 'photos-ng-api-client';

const instance: Media = {
    id,
    href,
    albumHref,
    capturedAt,
    type,
    filename,
    thumbnail,
    content,
    exif,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
