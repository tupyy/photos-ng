import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { mediaApi } from '@api/apiConfig';

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

// Real async thunk for uploading media files to the API
export const uploadMediaFiles = createAsyncThunk(
  'upload/uploadMediaFiles',
  async (payload: { files: { id: string; file: File }[]; albumId: string }, { dispatch, signal, rejectWithValue }) => {
    try {
      dispatch(uploadSlice.actions.setAlbumId(payload.albumId));
      const uploadedFiles: string[] = [];
      let uploadedCount = 0;

      // Upload each file individually
      for (const fileData of payload.files) {
        if (signal.aborted) {
          return rejectWithValue('Upload cancelled');
        }

        dispatch(uploadSlice.actions.updateFileStatus({ id: fileData.id, status: 'uploading' }));

        try {
          // Upload file to the API
          const response = await mediaApi.uploadMedia(
            fileData.file.name,
            payload.albumId,
            fileData.file
          );

          // Mark as completed
          dispatch(uploadSlice.actions.updateFileProgress({ id: fileData.id, progress: 100 }));
          dispatch(uploadSlice.actions.updateFileStatus({ id: fileData.id, status: 'completed' }));
          
          uploadedFiles.push(response.data.id);
          uploadedCount++;
        } catch (error) {
          console.error(`Failed to upload file ${fileData.file.name}:`, error);
          dispatch(uploadSlice.actions.updateFileStatus({ 
            id: fileData.id, 
            status: 'error',
            error: error instanceof Error ? error.message : 'Upload failed'
          }));
        }
      }

      return { success: true, uploadedCount, uploadedFiles };
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