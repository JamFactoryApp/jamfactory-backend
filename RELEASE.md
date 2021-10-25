# JamFactory Backend Release Notes

* [v0.1.0 (Latest)](#v010)

## vx.x.x

### Features Added

* :sparkles: Add User support and store all user information persistent in the redis database. No longer save spotify token is the session.
  See [What are Users](./docs/documentation.md#what-is-a-user)
* :sparkles: Add User management to get, set and delete your current user.
* :sparkles: Keep track which members have joined a JamSession. Each member is linked to a User.
  See [What are Members](./docs/documentation.md#what-is-a-member)
* :sparkles: Add rights management for the joined members. See [Member Rights](./docs/documentation.md#member-rights). Currently, there is no way to remove members from the JamSession.
* :sparkles: Keep history of the played songs of a JamSession
* :sparkles: You can now export the history and queued songs of a JamSession to a Spotify Playlist

### Features Removed

* :boom: Remove support for IP and Session voting. Remove changing the voting type of a JamSession. Use new user
  identifiers for voting as a default.

### Bug Fixes

* :bug: Don't return ``jamSession not found`` if a user isn't a member of a JamSession and tries to leave his current
  JamSession
  using [``GET: /api/v1/jam/leave``](./docs/documentation.md#5-leave-the-jamsession-currently-joined-by-the-user). He
  already isn't a member of any JamSession, so the route should indicate that and return a success confirmation.

### API Changes

* :boom: Remove ``PUT: /api/v1/auth/current`` endpoint, as it is replaced with the user management
* :sparkles: Add [``GET: /api/v1/me``](./docs/documentation.md#1-get-the-current-user-information) to get the current user information and authorization status.
* :sparkles: Add [``PUT: /api/v1/me``](./docs/documentation.md#2-set-the-current-user-information) to set the current user information.
* :sparkles: Add [``DELETE: /api/v1/me``](./docs/documentation.md#3-delete-the-current-user-information) to delete the current user information.
* :sparkles: Add [``GET: /api/v1/jam/members``](./docs/documentation.md#8-get-the-members-of-the-jamsession-joined-by-the-user) to get the joined members of the JamSession.
* :sparkles: Add [``SET: /api/v1/jam/members``](./docs/documentation.md#9-set-the-members-of-the-jamsession-joined-by-the-user) to set the joined members of the JamSession.
* :boom: Remove ``voting_type`` key from [``GET: /api/v1/jam``](./docs/documentation.md#2-get-the-information-of-the-jamsession-joined-by-the-user)
  response.
* :boom: Remove ``voting_type`` key from [``PUT: /api/v1/jam``](./docs/documentation.md#2-get-the-information-of-the-jamsession-joined-by-the-user)
  request and response.
* :sparkles: Add [``GET: /api/v1/queue/history``](./docs/documentation.md#5-get-the-played-song-history-of-the-jamsession-joined-by-the-user) to get the history of the JamSession.
* :sparkles: Add [``PUT: /api/v1/queue/export``](./docs/documentation.md#6-export-the-queue-to-a-playlist) to export the queue to a Playlist.

### Websocket Changes

* :sparkles: Add new [Websocket Event ``jam``](./docs/documentation.md#event-jam) sending changes of the JamSession information.
* :sparkles: Add new [Websocket Event ``members``](./docs/documentation.md#event-members) sending changes of the joined JamSession members.

## v0.1.0

:sparkles: This is the initial release of the JamFactory Backend project!
See [v0.1.0 Documentation](./docs/documentation.md)

### Features Added

* :sparkles: First stable release of the JamFactory backend.
* :sparkles: Handle user sessions using Redis. See [User Types](./docs/documentation.md#user-types)
* :sparkles: Login to Spotify and create a Spotify JamSession as a host.
* :sparkles: Join a created JamSession using the JamLabel.
* :sparkles: Search for Spotify Songs, Albums or Playlists.
* :sparkles: Save Spotify search results in a Redis cache.
* :sparkles: Add, vote for or delete songs, which are stored in a Queue.
  See [How voting works](./docs/documentation.md#how-voting-works)
* :sparkles: Add a Conductor to manage the JamSession and control the Spotify playback.
* :sparkles: Get the users Spotify Playlists and available Playback devices.
* :sparkles: Basic Playback control (Play, Pause and select the playback device)
* :sparkles: Detect, when a user takes Control over the Spotify playback and stop the Conductor.
  See [JamSession State](./docs/documentation.md#jamsession-state)
* :sparkles: Add Housekeeping to delete inactive JamSessions.
* :sparkles: Add Websocket Support to push updates to the joined users of a JamSession.
  See [Websocket Reference](./docs/documentation.md#socket-reference) for details.
* :memo: Added [API Documentation](./docs/documentation.md#api-reference)

### Features Removed

This is the first release. Nothing to remove.

### Bug Fixes

No bugs were fixed with this release.

### Added API Endpoints

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

* [``PUT: /api/v1/spotify/devices``](./docs/documentation.md#1-get-the-users-available-spotify-playback-devices)
* [``GET: /api/v1/spotify/playlists``](./docs/documentation.md#2-get-the-users-available-spotify-playlists)
* [``PUT: /api/v1/spotify/search``](./docs/documentation.md#3-search-for-an-item-on-spotify)

### Removed API Endpoints

No API Endpoints were removed in this release.
