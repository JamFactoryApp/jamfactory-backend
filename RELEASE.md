# JamFactory Backend Release Notes
### V0.1.0
[V0.1.0 Documentation](./docs/documentation.md)

#### Features Added
* First stable release of the JamFactory backend.
* Handle user sessions using Redis. See [User Types](./docs/documentation.md#user-types)
* Login to Spotify and create a Spotify JamSession as a host.
* Join a created JamSession using the JamLabel.
* Search for Spotify Songs, Albums or Playlists.
* Save Spotify search results in a Redis cache.
* Add, vote for or delete songs, which are stored in a Queue. See [How voting works](./docs/documentation.md#how-voting-works)
* Add a Conductor to manage the JamSession and control the Spotify playback. 
* Get the users Spotify Playlists and available Playback devices.
* Basic Playback control (Play, Pause and select the playback device)
* Detect, when a user takes Control over the Spotify playback and stop the Conductor. See [JamSession State](./docs/documentation.md#jamsession-state)
* Add Housekeeping to delete inactive JamSessions.
* Add Websocket Support to push updates to the joined users of a JamSession. See [Websocket Reference](./docs/documentation.md#socket-reference) for details.
* Added [API Documentation](./docs/documentation.md#api-reference)

#### Features Removed
This is the first release. Nothing to remove.


#### Bug Fixes
No bugs were fixed with this release


#### Added API Endpoints

**Authorization**
* [``GET: /api/v1/auth/current``](./docs/documentation.md#1-get-the-users-authorization-status)
* [``GET: /api/v1/auth/logout``](./docs/documentation.md#2-user-logout)
* [``GET: /api/v1/auth/login``](./docs/documentation.md#3-start-spotify-authorization-flow-for-user)

**JamSession**
* [``GET: /api/v1/jam/create``](./docs/documentation.md#1-create-a-new-jamsession)
* [``GET: /api/v1/jam``](./docs/documentation.md#2-get-the-information-of-the-jamsession-joined-by-the-user)
* [``PUT: /api/v1/jam/playback``](./docs/documentation.md#3-get-the-playback-of-the-jamsession-joined-by-the-user)
* [``PUT: /api/v1/jam/join``](./docs/documentation.md#4-join-an-existing-jamsession)
* [``GET: /api/v1/jam/leave``](./docs/documentation.md#5-leave-the-jamsession-currently-joined-by-the-user)
* [``PUT: /api/v1/jam/playback``](./docs/documentation.md#6-set-playback-of-the-jamsession-joined-by-the-user)
* [``PUT: /api/v1/jam``](./docs/documentation.md#7-set-the-information-of-the-jamsession-joined-by-the-user)

**Queue**
* [``PUT: /api/v1/queue/collection``](./docs/documentation.md#1-add-a-collection-to-the-queue-of-the-jamsession-joined-by-the-user)
* [``DELETE: /api/v1/delete``](./docs/documentation.md#2-delete-a-song-in-the-queue-of-the-jamsession-joined-by-the-user)
* [``GET: /api/v1/queue``](./docs/documentation.md#3-get-the-queue-of-the-jamsession-joined-by-the-user)
* [``PUT: /api/v1/queue/vote``](./docs/documentation.md#4-vote-for-a-song-in-the-queue-of-the-jamsession-joined-by-the-user)

**Spotify**
* [``PUT: /api/v1/spoity/devices``](./docs/documentation.md#1-get-the-users-available-spotify-playback-devices)
* [``GET: /api/v1/spoity/playlists``](./docs/documentation.md#2-get-the-users-available-spotify-playlists)
* [``PUT: /api/v1/spoity/search``](./docs/documentation.md#3-search-for-an-item-on-spotify)


####Removed API Endpoints
No API Endpoints will be removed in this release.