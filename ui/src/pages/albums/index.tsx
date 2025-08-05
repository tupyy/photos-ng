import React, { useEffect } from 'react';
import { useAppDispatch } from '@shared/store';
import { setPageActive } from '@shared/reducers/albumsSlice';
import AlbumsList from './components/AlbumsList';

const AlbumsPage: React.FC = () => {
  const dispatch = useAppDispatch();

  useEffect(() => {
    // Set page as active when component mounts
    dispatch(setPageActive(true));

    // Set page as inactive when component unmounts
    return () => {
      dispatch(setPageActive(false));
    };
  }, [dispatch]);

  return (
    <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <div className="px-4 py-6 sm:px-0">
        <AlbumsList />
      </div>
    </div>
  );
};

export default AlbumsPage;