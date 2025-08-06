import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';

export interface UploadFile {
  id: string;
  name: string;
  size: number;
  type: string;
  progress: number;
  status: 'pending' | 'uploading' | 'completed' | 'error';
  error?: string;
}

interface UploadState {
  files: UploadFile[];
  isUploading: boolean;
  albumId: string | null;
  error: string | null;
}

const initialState: UploadState = {
  files: [],
  isUploading: false,
  albumId: null,
  error: null,
};

// Stub async thunk for uploading media files
export const uploadMediaFiles = createAsyncThunk(
  'upload/uploadMediaFiles',
  async (payload: { fileIds: string[]; albumId: string }, { dispatch, signal, rejectWithValue, getState }) => {
    try {
      dispatch(uploadSlice.actions.setAlbumId(payload.albumId));

      // Simulate upload progress for each file
      for (const fileId of payload.fileIds) {
        if (signal.aborted) {
          return rejectWithValue('Upload cancelled');
        }

        dispatch(uploadSlice.actions.updateFileStatus({ id: fileId, status: 'uploading' }));

        // Simulate upload progress
        for (let progress = 0; progress <= 100; progress += 10) {
          if (signal.aborted) {
            return rejectWithValue('Upload cancelled');
          }
          
          dispatch(uploadSlice.actions.updateFileProgress({ id: fileId, progress }));
          await new Promise(resolve => setTimeout(resolve, 200)); // Simulate network delay
        }

        dispatch(uploadSlice.actions.updateFileStatus({ id: fileId, status: 'completed' }));
      }

      return { success: true, uploadedCount: payload.fileIds.length };
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Upload failed');
    }
  }
);

const uploadSlice = createSlice({
  name: 'upload',
  initialState,
  reducers: {
    setFiles: (state, action: PayloadAction<UploadFile[]>) => {
      state.files = action.payload;
    },
    setAlbumId: (state, action: PayloadAction<string>) => {
      state.albumId = action.payload;
    },
    addFiles: (state, action: PayloadAction<{ id: string; name: string; size: number; type: string }[]>) => {
      const newFiles: UploadFile[] = action.payload.map((fileData) => ({
        id: fileData.id,
        name: fileData.name,
        size: fileData.size,
        type: fileData.type,
        progress: 0,
        status: 'pending',
      }));
      state.files.push(...newFiles);
    },
    removeFile: (state, action: PayloadAction<string>) => {
      state.files = state.files.filter(file => file.id !== action.payload);
    },
    updateFileProgress: (state, action: PayloadAction<{ id: string; progress: number }>) => {
      const file = state.files.find(f => f.id === action.payload.id);
      if (file) {
        file.progress = action.payload.progress;
      }
    },
    updateFileStatus: (state, action: PayloadAction<{ id: string; status: UploadFile['status']; error?: string }>) => {
      const file = state.files.find(f => f.id === action.payload.id);
      if (file) {
        file.status = action.payload.status;
        if (action.payload.error) {
          file.error = action.payload.error;
        }
      }
    },
    clearFiles: (state) => {
      state.files = [];
      state.error = null;
    },
    clearError: (state) => {
      state.error = null;
    },
    reset: () => initialState,
  },
  extraReducers: (builder) => {
    builder
      .addCase(uploadMediaFiles.pending, (state) => {
        state.isUploading = true;
        state.error = null;
      })
      .addCase(uploadMediaFiles.fulfilled, (state) => {
        state.isUploading = false;
      })
      .addCase(uploadMediaFiles.rejected, (state, action) => {
        state.isUploading = false;
        state.error = action.payload as string || 'Upload failed';
      });
  },
});

export const {
  setFiles,
  setAlbumId,
  addFiles,
  removeFile,
  updateFileProgress,
  updateFileStatus,
  clearFiles,
  clearError,
  reset,
} = uploadSlice.actions;

export default uploadSlice.reducer;