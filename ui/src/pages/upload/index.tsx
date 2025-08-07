import React, { useState, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector, selectUpload, selectCurrentAlbum } from '@shared/store';
import { uploadMediaFiles, addFiles, removeFile, clearFiles, removeCompletedFiles, reset } from '@shared/reducers/uploadSlice';

const UploadMediaPage: React.FC = () => {
  const { albumId } = useParams<{ albumId: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const currentAlbum = useAppSelector(selectCurrentAlbum);
  const { files, isUploading, error } = useAppSelector(selectUpload);
  
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);
  const [actualFiles, setActualFiles] = useState<Record<string, File>>({});

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      dispatch(reset());
      // Clear actual files
      setActualFiles({});
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
      
      // Process each file
      for (let index = 0; index < validFiles.length; index++) {
        const file = validFiles[index];
        const fileId = `${timestamp}-${index}`;
        
        // Add to file data for Redux (serializable)
        fileDataArray.push({
          id: fileId,
          name: file.name,
          size: file.size,
          type: file.type,
        });
        
        // Store actual file object in component state
        newActualFiles[fileId] = file;
      }
      
      // Update Redux with serializable data
      dispatch(addFiles(fileDataArray));
      
      // Update component state with files
      setActualFiles(prev => ({ ...prev, ...newActualFiles }));
    }

    if (validFiles.length !== selectedFiles.length) {
      alert('Some files were not added. Only JPEG and PNG images are supported.');
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    handleFileSelect(e.target.files);
    // Reset input to allow selecting the same files again
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
    // Remove from actual files
    setActualFiles(prev => {
      const newFiles = { ...prev };
      delete newFiles[fileId];
      return newFiles;
    });
    
    dispatch(removeFile(fileId));
  };

  const handleUpload = async () => {
    if (!albumId || files.length === 0) return;

    try {
      // Prepare files with their actual File objects
      const filesToUpload = files.map(f => ({
        id: f.id,
        file: actualFiles[f.id]
      })).filter(f => f.file); // Filter out any missing files

      if (filesToUpload.length === 0) {
        console.error('No files available for upload');
        return;
      }

      await dispatch(uploadMediaFiles({ files: filesToUpload, albumId })).unwrap();
      
      // Navigate back to album after successful upload
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
    const pending = files.filter(f => f.status === 'pending').length;
    
    // Calculate overall progress percentage
    const totalProgress = files.reduce((acc, file) => {
      if (file.status === 'completed') return acc + 100;
      if (file.status === 'uploading') return acc + file.progress;
      return acc; // pending and error files contribute 0
    }, 0);
    
    const overallProgress = files.length > 0 ? Math.round(totalProgress / files.length) : 0;
    
    return { 
      completed, 
      failed, 
      uploading, 
      pending,
      total: files.length,
      overallProgress,
      currentlyUploading: uploading > 0 ? completed + 1 : completed // The file currently being uploaded
    };
  };

  const stats = getProgressStats();

  // Auto-remove completed files after a short delay
  useEffect(() => {
    if (stats.completed > 0 && !isUploading) {
      const timer = setTimeout(() => {
        dispatch(removeCompletedFiles());
      }, 2000); // Wait 2 seconds after upload completes before removing

      return () => clearTimeout(timer);
    }
  }, [stats.completed, isUploading, dispatch]);

  return (
    <div className="max-w-4xl mx-auto py-6 sm:px-6 lg:px-8">
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
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Upload Media</h1>
            {currentAlbum && (
              <p className="text-gray-600 dark:text-gray-400 mt-1">
                Upload photos to "{currentAlbum.name}"
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
              Supports JPEG and PNG images up to 10MB each
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
                    Uploading {stats.currentlyUploading} of {stats.total} files...
                  </p>
                  <p className="text-xs text-blue-700 dark:text-blue-200">
                    {stats.completed} completed, {stats.failed} failed
                  </p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-sm font-medium text-blue-900 dark:text-blue-100">
                  {stats.overallProgress}%
                </p>
              </div>
            </div>
            
            {/* Overall Progress Bar */}
            <div className="w-full bg-blue-200 dark:bg-blue-800 rounded-full h-2">
              <div
                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                style={{ width: `${stats.overallProgress}%` }}
              />
            </div>
          </div>
        )}

        {/* File List */}
        {files.length > 0 && (
          <div className="mt-6">
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
              Selected Files ({files.length})
            </h3>
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {files.map((file) => (
                <div
                  key={file.id}
                  className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-800 rounded-lg"
                >
                  <div className="flex items-center flex-1 min-w-0">
                    <div className="flex-shrink-0">
                      {file.type.startsWith('image/') ? (
                        <svg className="h-8 w-8 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
                        </svg>
                      ) : (
                        <svg className="h-8 w-8 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z" clipRule="evenodd" />
                        </svg>
                      )}
                    </div>
                    <div className="ml-4 flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                          {file.name}
                        </p>
                        <div className="flex items-center space-x-2">
                          <span className="text-xs text-gray-500 dark:text-gray-400">
                            {formatFileSize(file.size)}
                          </span>
                          {file.status === 'completed' && (
                            <svg className="w-4 h-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                            </svg>
                          )}
                          {file.status === 'error' && (
                            <svg className="w-4 h-4 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                            </svg>
                          )}
                        </div>
                      </div>
                      {file.status === 'uploading' && (
                        <div className="mt-2">
                          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                            <div
                              className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                              style={{ width: `${file.progress}%` }}
                            />
                          </div>
                          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                            {file.progress}% uploaded
                          </p>
                        </div>
                      )}
                      {file.status === 'error' && file.error && (
                        <p className="text-xs text-red-600 dark:text-red-400 mt-1">{file.error}</p>
                      )}
                    </div>
                  </div>
                  {!isUploading && (
                    <button
                      onClick={() => handleRemoveFile(file.id)}
                      className="ml-4 flex-shrink-0 p-1 text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
                    >
                      <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                      </svg>
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Error Display */}
        {error && (
          <div className="mt-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <div className="flex">
              <svg className="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800 dark:text-red-200">Upload Error</h3>
                <p className="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Action Buttons */}
        {files.length > 0 && (
          <div className="mt-6 flex justify-end space-x-3">
            {stats.completed > 0 && !isUploading && (
              <button
                onClick={() => dispatch(removeCompletedFiles())}
                className="px-4 py-2 border border-green-300 rounded-md text-sm font-medium text-green-700 bg-green-50 hover:bg-green-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 dark:bg-green-900/20 dark:border-green-600 dark:text-green-300 dark:hover:bg-green-900/30"
              >
                Remove Completed ({stats.completed})
              </button>
            )}
            <button
              onClick={() => dispatch(clearFiles())}
              disabled={isUploading}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
            >
              Clear All
            </button>
            <button
              onClick={handleUpload}
              disabled={isUploading || files.length === 0}
              className="px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isUploading ? `Uploading... (${stats.completed + stats.uploading}/${stats.total})` : `Upload ${files.length} File${files.length === 1 ? '' : 's'}`}
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default UploadMediaPage;