# Permissions

Datastore-level permissions for the user

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**canSync** | **string** | Whether the user can perform sync operations | [optional] [default to undefined]
**canCreateAlbums** | **string** | Whether the user can create new albums | [optional] [default to undefined]

## Example

```typescript
import { Permissions } from 'photos-ng-api-client';

const instance: Permissions = {
    canSync,
    canCreateAlbums,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
