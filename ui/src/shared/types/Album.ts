import { Album as GeneratedAlbum } from '@generated/models';

/**
 * Extended Album interface with full Album children instead of AlbumChildrenInner
 */
export interface Album extends Omit<GeneratedAlbum, 'children'> {
  children?: Album[];
}

export type { GeneratedAlbum as BaseAlbum };