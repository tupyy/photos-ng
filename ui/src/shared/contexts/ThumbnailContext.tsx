import React, { createContext, useContext, useState, ReactNode } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
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
  const location = useLocation();
  const { updateAlbum, fetchAlbums } = useAlbumsApi();
  
  const [isThumbnailMode, setIsThumbnailMode] = useState(false);
  const [thumbnailModeAlbumId, setThumbnailModeAlbumId] = useState<string | null>(null);
  const [startingLocation, setStartingLocation] = useState<string | null>(null);

  const startThumbnailSelection = (albumId: string) => {
    setThumbnailModeAlbumId(albumId);
    setIsThumbnailMode(true);
    setStartingLocation(location.pathname); // Store where the user started
    // Navigate to the album to start thumbnail selection
    navigate(`/albums/${albumId}`);
  };

  const exitThumbnailMode = () => {
    setIsThumbnailMode(false);
    setThumbnailModeAlbumId(null);
    // Navigate back to where the user started, or albums list as fallback
    navigate(startingLocation || '/albums');
    setStartingLocation(null);
  };

  const selectThumbnail = async (mediaId: string) => {
    if (!thumbnailModeAlbumId) return;
    
    const albumId = thumbnailModeAlbumId;
    const returnLocation = startingLocation;
    
    try {
      // Update album thumbnail
      await updateAlbum(albumId, { thumbnail: mediaId });
      
      // Exit thumbnail mode and navigate back to where the user started
      setIsThumbnailMode(false);
      setThumbnailModeAlbumId(null);
      setStartingLocation(null);
      navigate(returnLocation || '/albums');
      
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
