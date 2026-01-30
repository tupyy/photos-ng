import { useEffect, useCallback, useActionState } from 'react';
import { useAppDispatch, useAppSelector, selectCurrentAlbum } from '@shared/store';
import { createAlbum } from '@shared/reducers/albumsSlice';
import { CreateAlbumRequest } from '@generated/models';

export interface CreateAlbumFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: (albumId: string) => void;
}

interface FormState {
  error: string | null;
  success: boolean;
  albumId: string | null;
}

const initialState: FormState = {
  error: null,
  success: false,
  albumId: null,
};

const CreateAlbumForm = ({ isOpen, onClose, onSuccess }: CreateAlbumFormProps) => {
  const dispatch = useAppDispatch();
  const currentAlbum = useAppSelector(selectCurrentAlbum);

  const formAction = async (_prevState: FormState, formData: FormData): Promise<FormState> => {
    const name = formData.get('name') as string;
    const description = formData.get('description') as string;

    // Validation
    if (!name?.trim()) {
      return { error: 'Album name is required', success: false, albumId: null };
    }
    if (name.trim().length < 2) {
      return { error: 'Album name must be at least 2 characters', success: false, albumId: null };
    }

    try {
      const createRequest: CreateAlbumRequest = {
        name: name.trim(),
        ...(currentAlbum && { parentId: currentAlbum.id }),
        ...(description?.trim() && { description: description.trim() }),
      };

      const result = await dispatch(createAlbum(createRequest)).unwrap();
      return { error: null, success: true, albumId: result.id };
    } catch {
      return { error: 'Failed to create album. Please try again.', success: false, albumId: null };
    }
  };

  const [state, submitAction, isPending] = useActionState(formAction, initialState);

  // Handle success - close modal and call callback
  useEffect(() => {
    if (state.success && state.albumId) {
      onClose();
      onSuccess?.(state.albumId);
    }
  }, [state.success, state.albumId, onClose, onSuccess]);

  const handleCancel = useCallback(() => {
    onClose();
  }, [onClose]);

  // Handle ESC key to close modal
  useEffect(() => {
    const handleEscKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && isOpen && !isPending) {
        handleCancel();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscKey);
    }

    return () => {
      document.removeEventListener('keydown', handleEscKey);
    };
  }, [isOpen, isPending, handleCancel]);

  return (
    <div
      id="create-album-modal"
      tabIndex={-1}
      aria-hidden="true"
      className={`${isOpen ? 'flex' : 'hidden'} overflow-y-auto overflow-x-hidden fixed top-0 right-0 left-0 z-50 justify-center items-center w-full md:inset-0 h-[calc(100%-1rem)] max-h-full`}
    >
      <div className="relative p-4 w-full max-w-2xl max-h-full">
        {/* Modal content */}
        <div className="relative bg-white rounded-lg shadow-sm dark:bg-gray-700">
          {/* Modal header */}
          <div className="flex items-center justify-between p-4 md:p-5 border-b rounded-t dark:border-gray-600 border-gray-200">
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white">Create New Album</h3>
            <button
              type="button"
              onClick={handleCancel}
              className="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm w-8 h-8 ms-auto inline-flex justify-center items-center dark:hover:bg-gray-600 dark:hover:text-white"
            >
              <svg
                className="w-3 h-3"
                aria-hidden="true"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 14 14"
              >
                <path
                  stroke="currentColor"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
                />
              </svg>
              <span className="sr-only">Close modal</span>
            </button>
          </div>

          {/* Modal body */}
          <div className="p-4 md:p-5 space-y-4">
            <form action={submitAction} id="create-album-form">
              <div className="space-y-4">
                {/* Album Name */}
                <div>
                  <label htmlFor="name" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">
                    Name <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    id="name"
                    name="name"
                    className={`bg-gray-50 border text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500 ${
                      state.error ? 'border-red-500 dark:border-red-500' : 'border-gray-300 dark:border-gray-500'
                    }`}
                    placeholder="Enter album name"
                    disabled={isPending}
                    required
                  />
                  {state.error && <p className="mt-2 text-sm text-red-600 dark:text-red-500">{state.error}</p>}
                </div>

                {/* Album Description */}
                <div>
                  <label htmlFor="description" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">
                    Description
                  </label>
                  <textarea
                    id="description"
                    name="description"
                    rows={4}
                    className="block p-2.5 w-full text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                    placeholder="Enter album description (optional)"
                    disabled={isPending}
                  />
                </div>
              </div>
            </form>
          </div>

          {/* Modal footer */}
          <div className="flex items-center p-4 md:p-5 border-t border-gray-200 rounded-b dark:border-gray-600">
            <button
              type="submit"
              form="create-album-form"
              disabled={isPending}
              className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800 disabled:opacity-50 flex items-center"
            >
              {isPending ? (
                <>
                  <svg
                    className="animate-spin -ml-1 mr-3 h-4 w-4 text-white"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    ></circle>
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  Creating...
                </>
              ) : (
                'Create'
              )}
            </button>
            <button
              type="button"
              onClick={handleCancel}
              disabled={isPending}
              className="py-2.5 px-5 ms-3 text-sm font-medium text-gray-900 focus:outline-none bg-white rounded-lg border border-gray-200 hover:bg-gray-100 hover:text-blue-700 focus:z-10 focus:ring-4 focus:ring-gray-100 dark:focus:ring-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-600 dark:hover:text-white dark:hover:bg-gray-700 disabled:opacity-50"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CreateAlbumForm;
