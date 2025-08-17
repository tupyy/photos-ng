import React, { createContext, useContext, useState, ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAlbumsApi } from '@shared/hooks/useApi';

interface ThumbnailContextType {
  isThumbnailMode: boolean;
  thumbnailModeAlbumId: string | null;
  startThumbnailSelection: (albumId: string) => void;
  exitThumbnailMode: () => void;
  selectThumbnail: (mediaId: string) => Promise<void>;
}

const ThumbnailContext = createContext<ThumbnailContextType | undefined>(undefined);

interface ThumbnailProviderProps {
  children: ReactNode;
}

export const ThumbnailProvider: React.FC<ThumbnailProviderProps> = ({ children }) => {
  const navigate = useNavigate();
  const { updateAlbum, fetchAlbums } = useAlbumsApi();
  
  const [isThumbnailMode, setIsThumbnailMode] = useState(false);
  const [thumbnailModeAlbumId, setThumbnailModeAlbumId] = useState<string | null>(null);

  const startThumbnailSelection = (albumId: string) => {
    setThumbnailModeAlbumId(albumId);
    setIsThumbnailMode(true);
    // Navigate to the album to start thumbnail selection
    navigate(`/albums/${albumId}`);
  };

  const exitThumbnailMode = () => {
    setIsThumbnailMode(false);
    setThumbnailModeAlbumId(null);
    // Navigate back to albums list (where the user started)
    navigate('/albums');
  };

  const selectThumbnail = async (mediaId: string) => {
    if (!thumbnailModeAlbumId) return;
    
    try {
      // Update album thumbnail
      await updateAlbum(thumbnailModeAlbumId, { thumbnail: mediaId });
      
      // Exit thumbnail mode and navigate back to albums list
      setIsThumbnailMode(false);
      setThumbnailModeAlbumId(null);
      navigate('/albums');
      
      // Refresh albums to show new thumbnail
      fetchAlbums({ limit: 1000, offset: 0 });
    } catch (error) {
      console.error('Failed to set thumbnail:', error);
      throw error;
    }
  };

  const value: ThumbnailContextType = {
    isThumbnailMode,
    thumbnailModeAlbumId,
    startThumbnailSelection,
    exitThumbnailMode,
    selectThumbnail,
  };

  return (
    <ThumbnailContext.Provider value={value}>
      {children}
    </ThumbnailContext.Provider>
  );
};

export const useThumbnail = (): ThumbnailContextType => {
  const context = useContext(ThumbnailContext);
  if (context === undefined) {
    throw new Error('useThumbnail must be used within a ThumbnailProvider');
  }
  return context;
};
