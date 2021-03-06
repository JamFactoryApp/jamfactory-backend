# JamFactory

***V0.1.0 Documentation***

## Table of contents

* [Functional Documentation](#functional-documentation)
    * [Overview](#overview)
    * [User Types](#user-types)
    * [How voting works](#how-voting-works)
        * [Available voting types](#available-voting-types)
    * [JamSession State](#jamsession-state)
* [Object Model](#object-model)
    * [Queue Song](#queue-song)

* [API Reference](#api-reference)
    * [Authorization](#authorization)
        * [Get the User's Authorization Status](#1-get-the-users-authorization-status)
        * [User logout](#2-user-logout)
        * [Start Spotify Authorization Flow for User](#3-start-spotify-authorization-flow-for-user)
    * [JamSession](#jamsession)
        * [Create a new JamSession](#1-create-a-new-jamsession)
        * [Get the information of the JamSession joined by the user](#2-get-the-information-of-the-jamsession-joined-by-the-user)
        * [Get the playback of the JamSession joined by the user](#3-get-the-playback-of-the-jamsession-joined-by-the-user)
        * [Join an existing JamSession](#4-join-an-existing-jamsession)
        * [Leave the JamSession currently joined by the user](#5-leave-the-jamsession-currently-joined-by-the-user)
        * [Set playback of the JamSession joined by the user](#6-set-playback-of-the-jamsession-joined-by-the-user)
        * [Set the information of the JamSession joined by the user](#7-set-the-information-of-the-jamsession-joined-by-the-user)
    * [Queue](#queue)
        * [Add a collection to the queue of the JamSession joined by the user](#1-add-a-collection-to-the-queue-of-the-jamsession-joined-by-the-user)
        * [Delete a song in the queue of the JamSession joined by the user](#2-delete-a-song-in-the-queue-of-the-jamsession-joined-by-the-user)
        * [Get the queue of the JamSession joined by the user](#3-get-the-queue-of-the-jamsession-joined-by-the-user)
        * [Vote for a song in the queue of the JamSession joined by the user](#4-vote-for-a-song-in-the-queue-of-the-jamsession-joined-by-the-user)
    * [Spotify](#spotify)
        * [Get the User's Available Spotify Playback Devices](#1-get-the-users-available-spotify-playback-devices)
        * [Get the User's Available Spotify Playlists](#2-get-the-users-available-spotify-playlists)
        * [Search for an Item on Spotify](#3-search-for-an-item-on-spotify)

* [Socket Reference](#socket-reference)
    * [Socket Events](#socket-events)
        * [Event: ``queue`` ](#event-queue)
        * [Event: ``playback`` ](#event-playback)
        * [Event: ``close`` ](#event-close)
--------

## Functional Documentation

### Overview

JamFactory is an application which provides the necessary API to start a JamSession. 

A JamSession is a private party with **one** host to set it up, and many attendees to join in.

Your friends or party guests can vote for the songs they like and want to listen to, and the song with the most votes gets played next.
The host can select a Spotify playback device on which they wish to play the music.

The host of a JamSession has to have a Spotify premium account, the guests can join the JamSession without a Spotify account.

### User Types

The following user types are currently available:

| type      | description       
|----------	|-----------------
| ``New``   | The *User* did not join any *JamSession* as a *Guest* or created his own.
| ``Guest`` | The *User* joined an ongoing *JamSession* as a *Guest*. 	       
| ``Host``  | The *User* is logged into Spotify and started his own *JamSession* as a *Host*.

See the API description for information on the required user type for certain routes.

### How voting works

To decide which song will be played next in an ongoing JamSession, all guests, and the host, can vote for songs of their choice. Every
JamSession has a queue of songs, which is sorted by the number of votes per song. Which song will be played next, is based on the number of votes a song has. The more votes a song has,
the higher the song is in the queue. If it can't be clearly determined by votes, where the song should be placed in the queue, the order of the queue is dictated by the age of the songs added.
A song which has been longer in the queue will be placed higher than a more recently added song. When the currently played song ends, the song which is highest in the queue will be played next.
When a user votes for a song, which currently is not in the queue, the song will automatically be added to the queue with one initial vote of the user that suggested the song. 
If the song has already previously been added to the queue, that song's votes will be increased by one vote. A song can only exist once in a queue.

A vote can be retracted from a song in the queue, by voting again. Only the user's own vote can be taken away. When a song reaches zero votes, it will automatically be deleted
from the queue.

To keep track of which user voted for which song, each vote has a unique identifier based on the selected voting type of the JamSession. See
[Voting Types](#available-voting-types).

When a new JamSession is created, the queue will be empty. To get the jam going from the getgo, the host can decide to add a collection to the queue.
The songs in the collection are voted into the queue with one virtual vote. Although adding the collection to the queue, the host can still vote independently for each song of the collection. 

The virtual votes are added with a date in the future, meaning songs with only virtual votes will be added to the bottom of the queue, because "real" votes have a higher priority.
The songs with only virtual votes will serve as a fallback, as soon as there are no more songs with user votes left, so the jam never stops.

#### Available voting types

| type                  | description       
|----------	            |-----------------
| ``session_voting``    | The *Session ID* of the *User* is used as an identifier
| ``ip_voting``         | The *IP Address* of the *User* is used as an identifier	       


### JamSession State

To keep the jam going, each JamSession has its own conductor who will keep track of the current events. The conductor will, for instance,
check if the current song has ended, and if the next song should start. He also keeps track of the current playback of the session.

It is possible that the host does not want the conductor to interfere with the playback, even if the JamSession is still ongoing.
The JamSession state will determine if the conductor has the right to control the Spotify playback.

The following event will change the JamSession state:

| event                                                  | result       
|----------	                                             |-----------------
| *User* creates a *JamSession*                          | active = ``true`` if the user has a selected Playback Device, else inactive = ``false`` 
| *User* pauses playback through the *JamFactory App*    | inactive = ``false``
| *User* resumes playback through the *JamFactory App*   | active = ``true``
| *User* pauses playback through *Spotify*               | inactive = ``false``
| *User* resumes/starts playback through *Spotify*       | inactive = ``false``

## Object Model

### Queue Song

| type                      | description       
|----------	                |-----------------
| ``spotifyTrackFull``      | The *Session ID* of the *User* is used as an identifier
| ``votes``                 | The *IP Address* of the *User* is used as an identifier
| ``voted``                 | 

## API Reference

### Authorization

#### 1. Get the user's authorization status

***Description***

Get the user's current authorization status.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/auth/current
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type        | value description                                                                                          	    |
|----------	    |-----------------  |------------------------------------------------------------------------------------------------------------	    |
| ``user`` 	    | string 	        | Current *User* Type. See available [User Types](#user-types)                                      	            |
| ``label``     | string 	        | *JamLabel* of the currently joined *JamSession* for the *User*. Empty if the *User* has not joined a *JamSession*.|
| ``authorized``| boolean 	        | Current *Spotify* authorization status. ``True`` if the *User* completed the *Spotify* authorization process.     |

```json
{
    "user": "Host",
    "label": "GF7DZ",
    "authorized": true
}
```

#### 2. User logout


***Description***

Logout the current user. This will remove any session data of the user and remove the Spotify authorization. The JamSession will also be deleted if the user was its host. It is recommended to first leave any joined JamSession before logging out.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/auth/logout
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type        | value description                                                                                          	|
|----------	    |-----------------  |------------------------------------------------------------------------------------------------------------	|
| ``success`` 	| boolean 	        | Feedback if the logout process was successful.                                      	                        |

```json
{
    "success": true
}
```

#### 3. Start Spotify Authorization Flow for User

***Description***

Start the Spotify authorization process for the current user. Uses the [Authorization Code Flow](https://developer.spotify.com/documentation/general/guides/authorization-guide/#authorization-code-flow) from Spotify. Requires the user to have a Spotify premium account.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/auth/login
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type        | value description                                                                                          	|
|----------	    |-----------------  |------------------------------------------------------------------------------------------------------------	|
| ``url`` 	    | string 	        | The *Url* to the *Spotify Account Service* for authorization. See [Authorization Code Flow](https://developer.spotify.com/documentation/general/guides/authorization-guide/#authorization-code-flow)               	                        |

```json
{
    "url": "https://accounts.spotify.com/authorize"
}
```


### JamSession

#### 1. Create a new JamSession

***Description***

Create a new JamSession. Requires the user to be authorized by Spotify. The user will join the JamSession as the host.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/jam/create
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type            | value description                                     |
|----------	    |-------------------	|---------------------------------------------------    |
| ``label`` 	| string             	| The *JamLabel* of the created *JamSession*.                |

```json
{
    "label": "HG5FZ"
}
```

#### 2. Get the information of the JamSession joined by the user

***Description***

Get the information of the JamSession currently joined by the user. Requires the user to have joined a JamSession.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/jam
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	        | value type        	| value description                                                                                     |
|----------	        |-------------------	|-------------------------------------------------------------------------------------------------------|
| ``label``         | string  	            | *JamLabel* of the currently joined *JamSession*.                                       	            |
| ``name`` 	        | string             	| *Name* of the currently joined *JamSession*.    	                                                |
| ``active`` 	    | string            	| *State* of the currently joined *JamSession*. See [JamSession State](#jamsession-state)	|
| ``voting_type`` 	| string             	| *Voting type* of the currently joined *JamSession*. See [Available Voting Types](#available-voting-types)	|

```json
{
    "label": "TPMU4",
    "name": "Joe's Birthday Party",
    "active": true,
    "voting_type": "session_voting"
}
```

#### 3. Get the playback of the JamSession joined by the user

***Description***

Get the playback of the JamSession currently joined by the user. Requires the user to have joined a JamSession.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/jam/playback
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	        | value type        	                                                                                                                                        | value description                                     |
|----------	        |-------------------	                                                                                                                                        |---------------------------------------------------    |
| ``playback``      | [Spotify Playback Object](https://developer.spotify.com/documentation/web-api/reference-beta/#endpoint-get-information-about-the-users-current-playback)  	| *Playback state*                                      |
| ``device_id`` 	| string      	                                                                                                                                                | *Device ID* of the currently selected playback device.|

```json
{
    "playback": "<Spotify Playback Object>",
    "device_id": "abc123456"
}
```

#### 4. Join an existing JamSession

***Description***

Join an existing JamSession. The user will join the JamSession as a guest.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/jam/join
```

***Request Body (JSON):***

| key      	    | value type            | value description                                             |
|----------	    |-------------------	|---------------------------------------------------            |
| ``label`` 	| string *required*     | The *JamLabel* of the *JamSession* the *User* wants to join.  |

```json
{
	"label": "KWXBZ"
}
```

***Response Body (JSON):***

| key      	    | value type            | value description                                     |
|----------	    |-------------------	|---------------------------------------------------    |
| ``label`` 	| string             	| The *JamLabel* of the joined *JamSession*.            |

```json
{
	"label": "KWXBZ"
}
```

#### 5. Leave the JamSession currently joined by the user


***Description***

Leave the JamSession currently joined by the user. If the user is the host of the JamSession, the JamSession will be deleted. Requires the user to have joined a JamSession.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/jam/leave
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type            | value description                                     |
|----------	    |-------------------	|---------------------------------------------------    |
| ``success`` 	| boolean             	| Result of the operation.                              |

```json
{
	"success": true
}
```

#### 6. Set playback of the JamSession joined by the user


***Description***

Set the playback of the JamSession currently joined by the user. Requires the user to be the host of the JamSession.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/jam/playback
```

***Response Body (JSON):***

| key      	        | value type        	| value description                                     |
|----------	        |-------------------    |---------------------------------------------------    |
| ``device_id``     | string *optional*     | *Device ID* of the playback device.                   |
| ``playing`` 	    | boolean *optional*    | *Playback state*. ``True``= Play ``False`` = Pause    |

```json
{
    "device_id": "abc123456",
    "playing": false
}
```

***Response Body (JSON):***

| key      	        | value type        	                                                                                                                                        | value description                                     |
|----------	        |-------------------	                                                                                                                                        |---------------------------------------------------    |
| ``playback``      | [Spotify Playback Object](https://developer.spotify.com/documentation/web-api/reference-beta/#endpoint-get-information-about-the-users-current-playback)  	| Playback state                                     	|
| ``device_id`` 	| string             	                                                                                                                                        | *Device ID* of the currently selected playback device.|

```json
{
    "playback": "<Spotify Playback Object>",
    "device_id": "abc123456"
}
```


#### 7. Set the information of the JamSession joined by the user

***Description***

Get the information of the JamSession currently joined by the user. Requires the user to be the host of the JamSession.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/jam
```

***Request Body (JSON):***

| key      	        | value type        	| value description                                                                                                     |
|----------	        |-------------------	|-----------------------------------------------------------------------------------------------------------------------|
| ``name`` 	        | string *optional*    	| *Name* of the *JamSession* currently joined by the user.   	                                                        |
| ``active`` 	    | boolean *optional*    | *State* of the *JamSession* currently joined by the user. See [JamSession State](#jamsession-state)	                |
| ``voting_type`` 	| string *optional*    	| *Voting type* of the *JamSession* currently joined by the user. See [Available Voting Types](#available-voting-types)	|

```json
{
    "name": "Joe's Birthday Party",
    "active": true,
    "voting_type": "ip_voting"
}
```

***Response Body (JSON):***

| key      	        | value type        	| value description                                                                                                     |
|----------	        |-------------------	|-----------------------------------------------------------------------------------------------------------------------|
| ``label``         | string  	            | *JamLabel* of the *JamSession* currently joined by the user.                                       	                |
| ``name`` 	        | string             	| *Name* of the *JamSession* currently joined by the user.    	                                                        |
| ``active`` 	    | boolean            	| *State* of the *JamSession* currently joined by the user. See [JamSession State](#jamsession-state)	                |
| ``voting_type`` 	| string             	| *Voting type* of the *JamSession* currently joined by the user. See [Available Voting Types](#available-voting-types)	|

```json
{
    "label": "KWXBZ",
    "name": "Joe's Birthday Party",
    "active": true,
    "voting_type": "ip_voting"
}
```

### Queue

#### 1. Add a collection to the queue of the JamSession joined by the user

***Description***

Add a Spotify collection (playlist or album) to the current queue of the JamSession joined by the user. 
This can be used to add fallback music to a JamSession. The songs in the collection are voted in the queue with a virtual vote. 
The user adding the collection can still vote for the added songs. 
**Note** that the virtual votes, for the songs in the playlist, are added with a date in the future. A real vote of a user will overrule the virtual vote and therefore be listed higher up in the queue. 
For more details see [How voting works](#how-voting-works).

Requires the user to be the host of the JamSession.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/queue/collection
```

***Request Body (JSON):***

| key      	        | value type            | value description                                                                      |
|----------	        |-------------------	|----------------------------------------------------------------------------------------|
| ``collection`` 	| string *required*   	| *Spotify ID* of the collection.                                                        |
| ``type`` 	        | string *required*   	| *Type* of the collection. <br> ``playlist`` for a playlist <br> ``album`` for an album |

```json
{
	"collection": "3AGOiaoRXMSjswCLtuNqv5",
	"type": "album"
}
```

***Response Body (JSON):***

| key      	    | value type            | value description                                                                                                                                         |
|----------	    |-------------------	|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| ``queue`` 	| array             	| Array of the songs in the current *queue*. See [Queue Song](#queue-song). The *queue song objects* are personalized to the *user* requesting the *queue*. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

#### 2. Delete a song in the queue of the JamSession joined by the user


***Description***

Delete a song in the queue of the JamSession joined by the user. Only songs which are currently in the queue can be deleted. This deletes all existing votes for a song, but does not prevent a new vote for that song.

Requires the user to be the host of the JamSession.

***Endpoint:***

```bash
Method: DELETE
URL: jamfactory.app/api/v1/queue/delete
```

***Request Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``track`` 	| string *required*   	| *Spotify ID* of the track which should be deleted. See [Spotify Track Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#track-object-full)  |

```json
{
	"track": "2374M0fQpWi3dLnB54qaLX"
}
```

***Response Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``queue`` 	| array             	| Array of the songs in the current *queue*. See [Queue Song](#queue-song). The *queue song objects* are personalized to the *user* requesting the *queue*. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

#### 3. Get the queue of the JamSession joined by the user


***Description***

Returns the queue of the JamSession joined by the user. 
_**Don't query this endpoint on a regular basis.**_ 
To get updates on changes to the queue use the provided socket. See [Socket Reference](#socket-reference). 

Requires the user to have joined the JamSession.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/queue
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``queue`` 	| array             	| Array of the songs in the current *queue*. See [Queue Song](#queue-song). The *queue song objects* are personalized to the *user* requesting the *queue*. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

#### 4. Vote for a song in the queue of the JamSession joined by the user


***Description***

Add or remove a vote from the user to a song in the JamSession joined by the user. See [How voting works](#how-voting-works) for a more detailed description on how voting works.

Requires the user to have joined the JamSession.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/queue/vote
```

***Request Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``track`` 	| string *required*   	| *Spotify ID* of the track. See [Spotify Track Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#track-object-full)  |

```json
{
	"track": "2374M0fQpWi3dLnB54qaLX"
}
```

***Response Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``queue`` 	| array             	| Array of the songs in the current *queue*. See [Queue Song](#queue-song). The *queue song objects* are personalized to the *user* requesting the *queue*. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

### Spotify

#### 1. Get the User's available Spotify playback devices


***Description***

Get information about the current available devices of the user. Requires the current user to be a JamSession host.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/spotify/devices
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type        	| value description                                                                                          	|
|----------	    |-------------------	|------------------------------------------------------------------------------------------------------------	|
| ``devices``   | array 	            | Array of the available *Spotify playback devices* of the *User*. See [Spotify Device Object](https://developer.spotify.com/documentation/web-api/reference/player/get-a-users-available-devices/#device-object)                                    	|

```json
{
    "devices": "[]<Spotify Device Object>"
}
```

#### 2. Get the User's available Spotify playlists


***Description***

Get a list of all Spotify playlists owned or followed by the current user. Requires the current user to be a JamSession host.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/spotify/playlists
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	        | value type        	| value description                                                                                          	|
|----------	        |-------------------	|------------------------------------------------------------------------------------------------------------	|
| ``playlists`` 	| [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Array of the *users* available Spotify playlists as [Spotify Simplified Playlist Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#playlist-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object)

```json
{
    "playlists": "<Spotify Paging Object>"
}
```

#### 3. Search for an item on Spotify


***Description***

Get Spotify catalog information about playlists and tracks that match a keyword string. Requires the current user to have joined the JamSession.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/spotify/search
```

***Request Body (JSON):***

| key      	| value type        	| value description                                                                                          	|
|----------	|-------------------	|------------------------------------------------------------------------------------------------------------	|
| ``text`` 	| string *required* 	| Search text. All search texts are completed with a ``*`` for autofill.                                       	|
| ``type`` 	| string *required* 	| Type of the searched item. Available:<br>``track`` for Spotify tracks,<br>``playlist`` for Spotify playlists 	|

```json
{
	"text": "abba",
	"type": "track"
}
```

***Response Body (JSON):***

| key      	        | value type        	| value description                                                                                          	|
|----------	        |-------------------	|------------------------------------------------------------------------------------------------------------	|
| ``artists`` 	    | [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Spotify artists found with the submitted search term as [Spotify Simplified Artists Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#artist-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) |
| ``albums`` 	    | [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Spotify albums found with the submitted search term as [Spotify Simplified Album Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#album-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) |
| ``playlists`` 	| [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Spotify playlists found with the submitted search term as [Spotify Simplified Playlist Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#playlist-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) |
| ``tracks`` 	    | [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Spotify tracks found with the submitted search term as [Spotify Simplified Track Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#track-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) |

```json
{
    "artists": "<Spotify Paging Object>",
    "albums": "<Spotify Paging Object>",
    "playlists": "<Spotify Paging Object>",
    "tracks": "<Spotify Paging Object>"
}
```

# Socket Reference
JamFactory supports [Socket.IO](https://socket.io/) Websockets to update the user's information at certain events.
**Currently only Socket.IO Version <= 1.4 is supported**
When the user has joined a JamSession as a host or as a guest, he can connect to the Socket.IO room created by the JamSession.
The connection will automatically join the right room, based on the session cookie. 
The client only needs to open or close the connection and listen for the events.

***Example:***

Connect to the socket
```js
    import * as io from 'socket.io-client';
    this.socket = io.connect('http://jamfactory.app:3000');
```

Listen for events
```js
     this.socket.on('<EventName>', (message: any) => {
          // <Code to handle the message>
     });
```

Close the connection if the user leaves
```js
     this.socket.close();
```

## Socket events

### Event: ``queue`` 

The queue of the JamSession joined by the user has changed.

***Message (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``queue`` 	| array             	| Array of the songs in the current *queue*. See [Queue Song](#queue-song). The *queue song objects* are **not** personalized to the *user* requesting the *queue*. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

### Event: ``playback`` 

Update on the current playback state of the JamSession. This event is triggered approximately every second.

***Message (JSON):***

| key      	        | value type        	                                                                                            | value description                                     |
|----------	        |-------------------	                                                                                            |---------------------------------------------------    |
| ``playback``      | [Spotify Playback Object](https://developer.spotify.com/documentation/web-api/reference-beta/#endpoint-get-information-about-the-users-current-playback)  	| Playback state                                     	|
| ``device_id`` 	| string             	                                                                                            | *Device ID* of the currently selected playback device   |

```json
{
    "playback": "<Spotify Playback Object>",
    "device_id": "abc123456"
}
```

### Event: ``close`` 

The JamSession was closed.

***Message (String):***

Reason why the JamSession was closed.

| Reason      	    | Description       
|----------	        |-------------------
| ``host``          | The *host* closed the *JamSession*.
| ``inactive`` 	    | The *JamSession* was closed due to inactivity. 

---
[Back to top](#jamfactory)
> API Documentation Made with &#9829; by [thedevsaddam](https://github.com/thedevsaddam) | Generated at: 2020-10-03 13:28:49 by [docgen](https://github.com/thedevsaddam/docgen)
