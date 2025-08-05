import { useAppDispatch, useAppSelector, selectSync } from '@shared/store';
import { startSync, cancelSync, clearError } from '@reducers/index';

/**
 * Custom hook for managing sync operations
 */
export const useSync = () => {
  const dispatch = useAppDispatch();
  const syncState = useAppSelector(selectSync);

  const start = () => {
    return dispatch(startSync());
  };

  const cancel = () => {
    dispatch(cancelSync());
  };

  const clearSyncError = () => {
    dispatch(clearError());
  };

  return {
    ...syncState,
    start,
    cancel,
    clearError: clearSyncError,
  };
};

export default useSync;