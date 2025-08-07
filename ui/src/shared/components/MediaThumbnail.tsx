import { Media } from '@generated/models';

interface MediaThumbnailProps {
  media: Media;
  onInfoClick: (media: Media) => void;
}

const MediaThumbnail: React.FC<MediaThumbnailProps> = ({ media, onInfoClick }) => {
  const handleClick = () => {
    // TODO: Implement modal or lightbox for viewing full image
    console.log('Media clicked:', media);
  };

  const handleInfoClick = (e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent triggering the main click
    onInfoClick(media);
  };

  const handleError = (e: React.SyntheticEvent<HTMLImageElement>) => {
    // Fallback to a placeholder when thumbnail fails to load
    e.currentTarget.src =
      'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjNmNGY2Ii8+CiAgPHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5YTNhZiIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPk5vIEltYWdlPC90ZXh0Pgo8L3N2Zz4K';
  };

  return (
    <div
      className="relative aspect-square bg-gray-200 dark:bg-gray-700 rounded-lg overflow-hidden cursor-pointer hover:opacity-80 transition-opacity group"
      onClick={handleClick}
    >
      <img
        src={media.thumbnail}
        alt={`Media ${media.filename}`}
        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
        loading="lazy"
        onError={handleError}
      />

      {/* Info button */}
      <button
        onClick={handleInfoClick}
        className="absolute top-2 right-2 w-6 h-6 bg-black bg-opacity-50 hover:bg-opacity-70 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-200"
        title="View EXIF data"
      >
        <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
          <path
            fillRule="evenodd"
            d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
            clipRule="evenodd"
          />
        </svg>
      </button>
    </div>
  );
};

export default MediaThumbnail;
