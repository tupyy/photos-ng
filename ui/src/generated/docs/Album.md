# Album


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Unique identifier for the album | [default to undefined]
**href** | **string** |  | [default to undefined]
**name** | **string** | name of the album | [default to undefined]
**path** | **string** | path of the folder on disk | [default to undefined]
**parentHref** | **string** | href of the parent | [optional] [default to undefined]
**thumbnail** | **string** | href of the thumbnail | [optional] [default to undefined]
**description** | **string** |  | [optional] [default to undefined]
**children** | [**Array&lt;AlbumChildrenInner&gt;**](AlbumChildrenInner.md) |  | [optional] [default to undefined]
**mediaCount** | **number** | Total media including media of all its children | [default to undefined]
**media** | **Array&lt;string&gt;** | list of media href | [optional] [default to undefined]
**syncInProgress** | **boolean** | set true if a job syncing this album exists | [optional] [default to undefined]

## Example

```typescript
import { Album } from 'photos-ng-api-client';

const instance: Album = {
    id,
    href,
    name,
    path,
    parentHref,
    thumbnail,
    description,
    children,
    mediaCount,
    media,
    syncInProgress,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
