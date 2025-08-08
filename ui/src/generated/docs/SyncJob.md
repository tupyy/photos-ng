# SyncJob


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Unique identifier for the sync job | [default to undefined]
**status** | **string** | Current status of the sync job | [default to undefined]
**filesRemaining** | **number** | Number of files still to be processed | [default to undefined]
**totalFiles** | **number** | Total number of files to process | [default to undefined]
**filesProcessed** | [**Array&lt;ProcessedFile&gt;**](ProcessedFile.md) | List of processed files with their results | [default to undefined]

## Example

```typescript
import { SyncJob } from 'photos-ng-api-client';

const instance: SyncJob = {
    id,
    status,
    filesRemaining,
    totalFiles,
    filesProcessed,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
