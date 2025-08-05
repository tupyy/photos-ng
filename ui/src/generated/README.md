## photos-ng-api-client@1.0.0

This generator creates TypeScript/JavaScript client that utilizes [axios](https://github.com/axios/axios). The generated Node module can be used in the following environments:

Environment
* Node.js
* Webpack
* Browserify

Language level
* ES5 - you must have a Promises/A+ library installed
* ES6

Module system
* CommonJS
* ES6 module system

It can be used in both TypeScript and JavaScript. In TypeScript, the definition will be automatically resolved via `package.json`. ([Reference](https://www.typescriptlang.org/docs/handbook/declaration-files/consumption.html))

### Building

To build and compile the typescript sources to javascript use:
```
npm install
npm run build
```

### Publishing

First build the package then run `npm publish`

### Consuming

navigate to the folder of your consuming project and run one of the following commands.

_published:_

```
npm install photos-ng-api-client@1.0.0 --save
```

_unPublished (not recommended):_

```
npm install PATH_TO_GENERATED_PACKAGE --save
```

### Documentation for API Endpoints

All URIs are relative to *http://localhost:8080*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AlbumsApi* | [**createAlbum**](docs/AlbumsApi.md#createalbum) | **POST** /albums | Create a new album
*AlbumsApi* | [**deleteAlbum**](docs/AlbumsApi.md#deletealbum) | **DELETE** /albums/{id} | Delete album by ID
*AlbumsApi* | [**getAlbum**](docs/AlbumsApi.md#getalbum) | **GET** /albums/{id} | Get album by ID
*AlbumsApi* | [**listAlbums**](docs/AlbumsApi.md#listalbums) | **GET** /albums | List all albums
*AlbumsApi* | [**syncAlbum**](docs/AlbumsApi.md#syncalbum) | **POST** /albums/{id}/sync | Sync album
*AlbumsApi* | [**updateAlbum**](docs/AlbumsApi.md#updatealbum) | **PUT** /albums/{id} | Update album by ID
*MediaApi* | [**deleteMedia**](docs/MediaApi.md#deletemedia) | **DELETE** /media/{id} | Delete media by ID
*MediaApi* | [**getMedia**](docs/MediaApi.md#getmedia) | **GET** /media/{id} | Get media by ID
*MediaApi* | [**getMediaContent**](docs/MediaApi.md#getmediacontent) | **GET** /media/{id}/content | Get media content
*MediaApi* | [**getMediaThumbnail**](docs/MediaApi.md#getmediathumbnail) | **GET** /media/{id}/thumbnail | Get media thumbnail
*MediaApi* | [**listMedia**](docs/MediaApi.md#listmedia) | **GET** /media | List all media
*MediaApi* | [**updateMedia**](docs/MediaApi.md#updatemedia) | **PUT** /media/{id} | Update media by ID
*TimelineApi* | [**getTimeline**](docs/TimelineApi.md#gettimeline) | **GET** /timeline | Get timeline buckets


### Documentation For Models

 - [Album](docs/Album.md)
 - [AlbumChildrenInner](docs/AlbumChildrenInner.md)
 - [Bucket](docs/Bucket.md)
 - [CreateAlbumRequest](docs/CreateAlbumRequest.md)
 - [ExifHeader](docs/ExifHeader.md)
 - [GetTimelineResponse](docs/GetTimelineResponse.md)
 - [ListAlbumsResponse](docs/ListAlbumsResponse.md)
 - [ListMediaResponse](docs/ListMediaResponse.md)
 - [Media](docs/Media.md)
 - [ModelError](docs/ModelError.md)
 - [SyncAlbumResponse](docs/SyncAlbumResponse.md)
 - [UpdateAlbumRequest](docs/UpdateAlbumRequest.md)
 - [UpdateMediaRequest](docs/UpdateMediaRequest.md)


<a id="documentation-for-authorization"></a>
## Documentation For Authorization

Endpoints do not require authorization.

