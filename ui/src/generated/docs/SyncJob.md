# SyncJob


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Unique identifier for the sync job | [default to undefined]
**createdAt** | **string** |  | [default to undefined]
**startedAt** | **string** |  | [optional] [default to undefined]
**finishedAt** | **string** |  | [optional] [default to undefined]
**remainingTime** | **number** | aproximative ramaining running tile in seconds | [optional] [default to undefined]
**status** | **string** | Current status of the sync job | [default to undefined]
**remainingTasks** | **number** | Number of files still to be processed | [default to undefined]
**totalTasks** | **number** | Total number of files to process | [default to undefined]
**completedTasks** | [**Array&lt;TaskResult&gt;**](TaskResult.md) | List of processed files with their results | [default to undefined]

## Example

```typescript
import { SyncJob } from 'photos-ng-api-client';

const instance: SyncJob = {
    id,
    createdAt,
    startedAt,
    finishedAt,
    remainingTime,
    status,
    remainingTasks,
    totalTasks,
    completedTasks,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
