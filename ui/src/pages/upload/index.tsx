import { useState, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector, selectUpload, selectCurrentAlbum } from '@shared/store';
import { uploadMediaFiles, addFiles, removeFile, clearFiles, removeCompletedFiles, reset } from '@shared/reducers/uploadSlice';
import { CheckIcon, XMarkIcon } from '@heroicons/react/24/outline';

interface FileData {
  id: string;
  name: string;
  size: number;
  type: string;
  status: 'pending' | 'uploading' | 'completed' | 'error';
  progress: number;
  error?: string;
}

interface ThumbnailItemProps {
  file: FileData;
  thumbnailUrl: string | undefined;
  onRemove: () => void;
  disabled: boolean;
}

const ThumbnailItem = ({ file, thumbnailUrl, onRemove, disabled }: ThumbnailItemProps) => (
  <div className="relative aspect-square group">
    {thumbnailUrl ? (
      <img
        src={thumbnailUrl}
        alt={file.name}
        className="w-full h-full object-cover rounded-lg bg-gray-100 dark:bg-gray-800"
      />
    ) : (
      <div className="w-full h-full rounded-lg bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
        <svg className="w-6 h-6 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
        </svg>
      </div>
    )}

    {/* Uploading overlay */}
    {file.status === 'uploading' && (
      <div className="absolute inset-0 bg-black/50 rounded-lg flex items-center justify-center">
        <svg className="animate-spin w-6 h-6 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>
    )}

    {/* Completed checkmark */}
    {file.status === 'completed' && (
      <div className="absolute top-1 right-1 bg-green-500 rounded-full p-0.5">
        <CheckIcon className="w-3 h-3 text-white" />
      </div>
    )}

    {/* Error overlay */}
    {file.status === 'error' && (
      <div className="absolute inset-0 bg-red-500/50 rounded-lg flex items-center justify-center" title={file.error}>
        <XMarkIcon className="w-6 h-6 text-white" />
      </div>
    )}

    {/* Remove button on hover */}
    {!disabled && file.status !== 'completed' && file.status !== 'uploading' && (
      <button
        onClick={onRemove}
        className="absolute top-1 right-1 bg-black/60 hover:bg-black/80 rounded-full p-0.5 opacity-0 group-hover:opacity-100 transition-opacity"
      >
        <XMarkIcon className="w-3.5 h-3.5 text-white" />
      </button>
    )}
  </div>
);

const UploadMediaPage = () => {
  const { albumId } = useParams<{ albumId: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const currentAlbum = useAppSelector(selectCurrentAlbum);
  const { files, isUploading, error } = useAppSelector(selectUpload);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);
  const [actualFiles, setActualFiles] = useState<Record<string, File>>({});
  const [thumbnailUrls, setThumbnailUrls] = useState<Record<string, string>>({});

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      dispatch(reset());
      setActualFiles({});
      // Revoke all thumbnail URLs to free memory
      Object.values(thumbnailUrls).forEach(url => URL.revokeObjectURL(url));
    };
  }, [dispatch]);

  const handleFileSelect = async (selectedFiles: FileList | null) => {
    if (!selectedFiles) return;

    const validFiles = Array.from(selectedFiles).filter(file => {
      const isImage = file.type.startsWith('image/');
      const isValidFormat = ['image/jpeg', 'image/jpg', 'image/png'].includes(file.type);
      return isImage && isValidFormat;
    });

    if (validFiles.length > 0) {
      const timestamp = Date.now();
      const fileDataArray: { id: string; name: string; size: number; type: string }[] = [];
      const newActualFiles: Record<string, File> = {};
      const newThumbnailUrls: Record<string, string> = {};

      for (let index = 0; index < validFiles.length; index++) {
        const file = validFiles[index];
        const fileId = `${timestamp}-${index}`;

        fileDataArray.push({
          id: fileId,
          name: file.name,
          size: file.size,
          type: file.type,
        });

        newActualFiles[fileId] = file;
        newThumbnailUrls[fileId] = URL.createObjectURL(file);
      }

      dispatch(addFiles(fileDataArray));
      setActualFiles(prev => ({ ...prev, ...newActualFiles }));
      setThumbnailUrls(prev => ({ ...prev, ...newThumbnailUrls }));
    }

    if (validFiles.length !== selectedFiles.length) {
      alert('Some files were not added. Only JPEG and PNG images are supported.');
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    handleFileSelect(e.target.files);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleBrowseClick = () => {
    fileInputRef.current?.click();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    handleFileSelect(e.dataTransfer.files);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
  };

  const handleRemoveFile = (fileId: string) => {
    // Revoke thumbnail URL
    if (thumbnailUrls[fileId]) {
      URL.revokeObjectURL(thumbnailUrls[fileId]);
      setThumbnailUrls(prev => {
        const newUrls = { ...prev };
        delete newUrls[fileId];
        return newUrls;
      });
    }

    setActualFiles(prev => {
      const newFiles = { ...prev };
      delete newFiles[fileId];
      return newFiles;
    });

    dispatch(removeFile(fileId));
  };

  const handleClearAll = () => {
    // Revoke all thumbnail URLs
    Object.values(thumbnailUrls).forEach(url => URL.revokeObjectURL(url));
    setThumbnailUrls({});
    setActualFiles({});
    dispatch(clearFiles());
  };

  const handleUpload = async () => {
    if (!albumId || files.length === 0) return;

    try {
      const filesToUpload = files.map(f => ({
        id: f.id,
        file: actualFiles[f.id]
      })).filter(f => f.file);

      if (filesToUpload.length === 0) {
        console.error('No files available for upload');
        return;
      }

      await dispatch(uploadMediaFiles({ files: filesToUpload, albumId })).unwrap();

      setTimeout(() => {
        navigate(`/albums/${albumId}`);
      }, 1000);
    } catch (error) {
      console.error('Upload failed:', error);
    }
  };

  const handleCancel = () => {
    if (albumId) {
      navigate(`/albums/${albumId}`);
    } else {
      navigate('/albums');
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getProgressStats = () => {
    const completed = files.filter(f => f.status === 'completed').length;
    const failed = files.filter(f => f.status === 'error').length;
    const uploading = files.filter(f => f.status === 'uploading').length;

    const totalProgress = files.reduce((acc, file) => {
      if (file.status === 'completed') return acc + 100;
      if (file.status === 'uploading') return acc + file.progress;
      return acc;
    }, 0);

    const overallProgress = files.length > 0 ? Math.round(totalProgress / files.length) : 0;

    return {
      completed,
      failed,
      uploading,
      total: files.length,
      overallProgress,
      currentlyUploading: uploading > 0 ? completed + 1 : completed
    };
  };

  const stats = getProgressStats();
  const totalSize = files.reduce((acc, f) => acc + f.size, 0);

  // Auto-remove completed files after a short delay
  useEffect(() => {
    if (stats.completed > 0 && !isUploading) {
      const timer = setTimeout(() => {
        dispatch(removeCompletedFiles());
      }, 2000);

      return () => clearTimeout(timer);
    }
  }, [stats.completed, isUploading, dispatch]);

  // Dynamic page title
  const pageTitle = isUploading
    ? `Uploading ${stats.completed + 1}/${stats.total} - Photos`
    : currentAlbum
      ? `Upload to ${currentAlbum.name} - Photos`
      : 'Upload - Photos';

  return (
    <div className="max-w-4xl mx-auto py-6 sm:px-6 lg:px-8">
      <title>{pageTitle}</title>
      <div className="px-4 py-6 sm:px-0">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center justify-between">
            <button
              onClick={handleCancel}
              className="flex items-center text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" />
              </svg>
              Back to Album
            </button>
          </div>
          <div className="mt-4">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Upload Photos</h1>
            {currentAlbum && (
              <p className="text-gray-600 dark:text-gray-400 mt-1">
                to "{currentAlbum.name}"
              </p>
            )}
          </div>
        </div>

        {/* File Drop Zone */}
        <div
          className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
            dragOver
              ? 'border-blue-400 bg-blue-50 dark:bg-blue-900/20'
              : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'
          }`}
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
        >
          <svg
            className="mx-auto h-12 w-12 text-gray-400 dark:text-gray-500"
            stroke="currentColor"
            fill="none"
            viewBox="0 0 48 48"
          >
            <path
              d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <div className="mt-4">
            <p className="text-lg font-medium text-gray-900 dark:text-white">
              Drop your photos here, or{' '}
              <button
                type="button"
                onClick={handleBrowseClick}
                className="text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300"
              >
                browse
              </button>
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">
              JPEG and PNG images up to 10MB each
            </p>
          </div>
          <input
            ref={fileInputRef}
            type="file"
            multiple
            accept="image/jpeg,image/jpg,image/png"
            onChange={handleFileInputChange}
            className="hidden"
          />
        </div>

        {/* Upload Progress Stats */}
        {files.length > 0 && isUploading && (
          <div className="mt-6 bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center">
                <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <div>
                  <p className="text-sm font-medium text-blue-900 dark:text-blue-100">
                    Uploading {stats.currentlyUploading} of {stats.total}
                  </p>
                  {stats.failed > 0 && (
                    <p className="text-xs text-red-600 dark:text-red-400">
                      {stats.failed} failed
                    </p>
                  )}
                </div>
              </div>
              <p className="text-sm font-medium text-blue-900 dark:text-blue-100">
                {stats.overallProgress}%
              </p>
            </div>
            <div className="w-full bg-blue-200 dark:bg-blue-800 rounded-full h-2">
              <div
                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                style={{ width: `${stats.overallProgress}%` }}
              />
            </div>
          </div>
        )}

        {/* Error Display */}
        {error && (
          <div className="mt-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <div className="flex">
              <XMarkIcon className="h-5 w-5 text-red-400" />
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800 dark:text-red-200">Upload Error</h3>
                <p className="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Thumbnail Grid Section */}
        {files.length > 0 && (
          <div className="mt-6">
            {/* Action bar - above grid */}
            <div className="flex items-center justify-between mb-4">
              <p className="text-sm text-gray-600 dark:text-gray-400">
                {files.length} photo{files.length !== 1 ? 's' : ''} ({formatFileSize(totalSize)})
              </p>
              <div className="flex items-center gap-2">
                {!isUploading && (
                  <button
                    onClick={handleClearAll}
                    className="px-3 py-1.5 text-sm text-gray-600 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-400 transition-colors"
                  >
                    Clear
                  </button>
                )}
                <button
                  onClick={handleCancel}
                  disabled={isUploading}
                  className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleUpload}
                  disabled={isUploading || files.length === 0}
                  className="px-4 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded-md disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {isUploading ? 'Uploading...' : `Upload ${files.length}`}
                </button>
              </div>
            </div>

            {/* Thumbnail grid */}
            <div className="grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 gap-2">
              {files.map((file) => (
                <ThumbnailItem
                  key={file.id}
                  file={file as FileData}
                  thumbnailUrl={thumbnailUrls[file.id]}
                  onRemove={() => handleRemoveFile(file.id)}
                  disabled={isUploading}
                />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default UploadMediaPage;
