import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { UserApi } from '@generated/api/user-api';
import { User } from '@generated/models';
import { apiConfig } from '@shared/api/apiConfig';

export interface UserState {
  user: User | null;
  loading: boolean;
  error: string | null;
  initialized: boolean;
}

const initialState: UserState = {
  user: null,
  loading: false,
  error: null,
  initialized: false,
};

const userApi = new UserApi(apiConfig);

// Async thunk to fetch current user
export const fetchCurrentUser = createAsyncThunk(
  'user/fetchCurrentUser',
  async (_, { rejectWithValue }) => {
    try {
      const response = await userApi.getCurrentUser();
      return response.data;
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to fetch user'
      );
    }
  }
);

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    clearUser: (state) => {
      state.user = null;
      state.initialized = false;
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchCurrentUser.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchCurrentUser.fulfilled, (state, action) => {
        state.loading = false;
        state.user = action.payload;
        state.initialized = true;
        state.error = null;
      })
      .addCase(fetchCurrentUser.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
        state.initialized = true;
      });
  },
});

export const { clearUser } = userSlice.actions;
export default userSlice.reducer;
