import React, { useEffect } from 'react';
import { useAppDispatch, useAppSelector, selectAlbumsCreateFormOpen } from '@shared/store';
import { setPageActive, setCreateFormOpen } from '@shared/reducers/albumsSlice';
import AlbumsList from './components/AlbumsList';
import CreateAlbumForm from './components/CreateAlbumForm';

const AlbumsPage: React.FC = () => {
  const dispatch = useAppDispatch();
  const isCreateFormOpen = useAppSelector(selectAlbumsCreateFormOpen);

  useEffect(() => {
    // Set page as active when component mounts
    dispatch(setPageActive(true));

    // Set page as inactive when component unmounts
    return () => {
      dispatch(setPageActive(false));
    };
  }, [dispatch]);

  const handleCreateFormClose = () => {
    dispatch(setCreateFormOpen(false));
  };

  const handleCreateAlbumSuccess = (albumId: string) => {
    console.log('Album created successfully:', albumId);
    // TODO: Navigate to the created album or refresh the albums list
  };

  return (
    <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <div className="px-4 py-6 sm:px-0">
        <AlbumsList />
      </div>
      
      {/* Create Album Form Modal */}
      <CreateAlbumForm 
        isOpen={isCreateFormOpen}
        onClose={handleCreateFormClose}
        onSuccess={handleCreateAlbumSuccess}
      />
    </div>
  );
};

export default AlbumsPage;