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
        * [Logout the User](#2-logout-the-user)
        * [Start Spotify Authorization Flow for User](#3-start-spotify-authorization-flow-for-user)
    * [JamSession](#jamsession)
        * [Create a new JamSession](#1-create-a-new-jamsession)
        * [Get the Information of the User's joined JamSession](#2-get-the-information-of-the-users-joined-jamsession)
        * [Get the Playback of the User's joined JamSession](#3-get-the-playback-of-the-users-joined-jamsession)
        * [Join an existing JamSession](#4-join-an-existing-jamsession)
        * [Leave the User's currently joined JamSession](#5-leave-the-users-currently-joined-jamsession)
        * [Set Playback of the current User's JamSession](#6-set-playback-of-the-current-users-jamsession)
        * [Set the Information of the User's joined JamSession](#7-set-the-information-of-the-users-joined-jamsession)
    * [Queue](#queue)
        * [Add a Collection to the Queue of the User's joined JamSession](#1-add-a-collection-to-the-queue-of-the-users-joined-jamsession)
        * [Delete a Song in the Queue of the User's joined JamSession](#2-delete-a-song-in-the-queue-of-the-users-joined-jamsession)
        * [Get the Queue of the User's joined JamSession](#3-get-the-queue-of-the-users-joined-jamsession)
        * [Vote for a Song in the Queue of the User's joined JamSession](#4-vote-for-a-song-in-the-queue-of-the-users-joined-jamsession)
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

JamFactory is a Application which provides the necessary API start a JamSession with your friends. Let your friends vote for the Music 
your selected Spotify Playback Device will play. Only the Host of a JamSession needs a Spotify Premium Account.

### User Types

The following user types are currently available:

| type      | description       
|----------	|-----------------
| ``New``   | The User did not join any JamSession as a Guest or created his own.
| ``Guest`` | The User joined an ongoing JamSession as a Guest. 	       
| ``Host``  | The User is logged into Spotify and started his own JamSession as a Host.

See API Description for information about the required User Type for certain routes.

### How voting works

To decide what is played next in ongoing JamSession all Guests and the Host are able to Vote Songs of there choice. Each
JamSession has a Queue of Songs which are played next based on the number of Votes each Song has. When a User votes for a Song,
which currently is not in the Queue, the Song will be added to the Queue with the Users initial Vote. If the Song already exists the Vote
of the User will be added to the Song, resulting in a higher place in the Queue.

If a User votes for the same Song again, the Users Vote will be removed from the Song. If a Song reaches zero Votes it will be deleted
from the Queue.

When Song, which is currently playing, ends the Song with the most Votes is played next. To determine which Song has the most Votes all
Songs in the Queue are ordered by their number of Votes. If Songs have the same number of Votes the older Song will be placed higer in the Queue.
The first Song in the orderd Queue will be played as the next Song.

To keep Track of which User voted for which Song each Vote has a unique identifier based on the selected voting type of the Jam Session. See
[Voting Types](#available-voting-types).

When starting a new JamSession the Queue can be quite empty. To get the Jam going the Host can decide to add a Collection to the Queue.
The Songs in the Collection are voted in the queue with a virtual Vote. The User adding the Collection can still Vote for the added Songs. 
However the virtual Votes for the songs in the Collection are added with a date in the future. Therefore a real Vote of a User always 
overrules the virtual Vote. If no Songs with real Votes are left the Songs from the added Collection will serve as a Fallback.  

#### Available voting types

| type                  | description       
|----------	            |-----------------
| ``session_voting``    | The Session ID of the User is used as an identifier
| ``ip_voting``         | The IP Address of the User is used as an identifier	       


### JamSession State

To keep the music coming each JamSession has a Conductor who will keep track of the current events. For example the Conductor will
check if the current Song ended and the next Song should start. He also keeps track of the current Playback of the Session.

Sometimes the Host of the User does not want to Conductor to interfere with his Spotify Playback even if the JamSession is still going.
The JamSession state will determine if the Conductor has the instruction to control the Spotify Playback.

The following event will change the JamSession State

| event                                              | result       
|----------	                                         |-----------------
| User creates a JamSession                          | active = ``true`` if the user has a selected Playback Device, else inactive = ``false`` 
| User pauses playback through the JamFactory App    | inactive = ``false``
| User resumes playback through the JamFactory App   | active = ``true``
| User pauses playback through Spotify               | inactive = ``false``
| User resumes/starts playback through Spotify       | inactive = ``false``

## Object Model

### Queue Song

| type                      | description       
|----------	                |-----------------
| ``spotifyTrackFull``      | The Session ID of the User is used as an identifier
| ``votes``                 | The IP Adress of the User is used as an identifier
| ``voted``                 | 

## API Reference

### Authorization

#### 1. Get the User's Authorization Status

***Description***

Get the current User's authorization status.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/auth/current
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type        | value description                                                                                          	|
|----------	    |-----------------  |------------------------------------------------------------------------------------------------------------	|
| ``user`` 	    | string 	        | Current User Type. See available [User Types](#user-types)                                      	            |
| ``label``     | string 	        | JamLabel of the User's currently joined JamSession. Empty if the User is not joined a JamSession              |
| ``authorized``| boolean 	        | Current Spotify authorization Status. True if the User completed the Spotify authorization process 	        |

```json
{
    "user": "Host",
    "label": "GF7DZ",
    "authorized": true
}
```

#### 2. Logout the User


***Description***

Logout the current User. This will remove any session data of the user and remove the spotify authorization. If the User was a Host of a JamSession then the Jam session will also be deleted. It is recommended to first leave any currently joined JamSession before logout.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/auth/logout
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type        | value description                                                                                          	|
|----------	    |-----------------  |------------------------------------------------------------------------------------------------------------	|
| ``success`` 	| boolean 	        | Feedback if the Logout process was successfull.                                      	                        |

```json
{
    "success": true
}
```

#### 3. Start Spotify Authorization Flow for User

***Description***

Start the Spotify Authorization Process for the current User. Uses the [Authorization Code Flow](https://developer.spotify.com/documentation/general/guides/authorization-guide/#authorization-code-flow) from spotify. Required the User to have a Spotify Premium Account.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/auth/login
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type        | value description                                                                                          	|
|----------	    |-----------------  |------------------------------------------------------------------------------------------------------------	|
| ``url`` 	    | string 	        | The Url to the Spotify Account Service for the Authorization. See [Authorization Code Flow](https://developer.spotify.com/documentation/general/guides/authorization-guide/#authorization-code-flow)               	                        |

```json
{
    "url": "https://accounts.spotify.com/authorize"
}
```


### JamSession

#### 1. Create a new JamSession

***Description***

Create a new JamSession. Requires the User to be authorized by Spotify. The user will join the JamSession as Host.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/jam/create
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type            | value description                                     |
|----------	    |-------------------	|---------------------------------------------------    |
| ``label`` 	| string             	| The JamLabel of the created JamSession                |

```json
{
    "label": "HG5FZ"
}
```

#### 2. Get the Information of the User's joined JamSession

***Description***

Get the Information of the User's currently joined JamSession. Requires the User to be joined a JamSession

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/jam
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	        | value type        	| value description                                                                                     |
|----------	        |-------------------	|-------------------------------------------------------------------------------------------------------|
| ``label``         | string  	            | JamLabel of the User's currently joined JamSession                                       	            |
| ``name`` 	        | string             	| Name of the User's currently joined JamSession    	                                                |
| ``active`` 	    | string            	| State of the User's currently joined JamSession. See [JamSession State](#jamsession-state)	|
| ``voting_type`` 	| string             	| Voting type of the User's currently joined JamSession. See [Avaliable Voting Types](#available-voting-types)	|

```json
{
    "label": "TPMU4",
    "name": "Joe's Birthday Party",
    "active": true,
    "voting_type": "session_voting"
}
```

#### 3. Get the Playback of the User's joined JamSession

***Description***

Get the Playback of the User's currently joined JamSession. Requires the User to be joined a JamSession

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/jam/playback
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	        | value type        	                                                                                            | value description                                     |
|----------	        |-------------------	                                                                                            |---------------------------------------------------    |
| ``playback``      | [Spotify Playback Object](https://developer.spotify.com/documentation/web-api/reference-beta/#endpoint-get-information-about-the-users-current-playback)  	| Playback state                                     	|
| ``device_id`` 	| string      	                                                                                            | Device ID of the currently selected playback device   |

```json
{
    "playback": "<Spotify Playback Object>",
    "device_id": "abc123456"
}
```

#### 4. Join an existing JamSession

***Description***

Join an existing JamSession. The user will join the JamSession as a Guest.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/jam/join
```

***Request Body (JSON):***

| key      	    | value type            | value description                                     |
|----------	    |-------------------	|---------------------------------------------------    |
| ``label`` 	| string *required*           	| The JamLable of JamSession the user wants to join              |

```json
{
	"label": "KWXBZ"
}
```

***Response Body (JSON):***

| key      	    | value type            | value description                                     |
|----------	    |-------------------	|---------------------------------------------------    |
| ``label`` 	| string             	| The JamLabel of the joined JamSession                 |

```json
{
	"label": "KWXBZ"
}
```

#### 5. Leave the User's currently joined JamSession


***Description***

Leave the User's currently joined JamSession. If the User is the Host of the JamSession the JamSession will be deleted. Requires the User to be joined an JamSession.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/jam/leave
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type            | value description                                     |
|----------	    |-------------------	|---------------------------------------------------    |
| ``success`` 	| boolean             	| Result of the operation                               |

```json
{
	"success": true
}
```

#### 6. Set Playback of the current User's JamSession


***Description***

Set the Playback of the User's currently joined JamSession. Requires the User to be the Host of a JamSession

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/jam/playback
```

***Response Body (JSON):***

| key      	        | value type        	| value description                                     |
|----------	        |-------------------    |---------------------------------------------------    |
| ``device_id``     | string *optional*     | Device ID of the playback device                      |
| ``playing`` 	    | boolean *optional*    | Playback state. ``True``= Play ``False`` = Pause      |

```json
{
    "device_id": "abc123456",
    "playing": false
}
```

***Response Body (JSON):***

| key      	        | value type        	                                                                                            | value description                                     |
|----------	        |-------------------	                                                                                            |---------------------------------------------------    |
| ``playback``      | [Spotify Playback Object](https://developer.spotify.com/documentation/web-api/reference-beta/#endpoint-get-information-about-the-users-current-playback)  	| Playback state                                     	|
| ``device_id`` 	| string             	                                                                                            | Device ID of the currently selected playback device   |

```json
{
    "playback": "<Spotify Playback Object>",
    "device_id": "abc123456"
}
```


#### 7. Set the Information of the User's joined JamSession

***Description***

Get the Information of the User's currently joined JamSession. Requires the User to be the Host of a JamSession

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/jam
```

***Request Body (JSON):***

| key      	        | value type        	| value description                                                                                     |
|----------	        |-------------------	|-------------------------------------------------------------------------------------------------------|
| ``name`` 	        | string *optional*    	| Name of the User's currently joined JamSession    	                                                |
| ``active`` 	    | boolean *optional*    | State of the User's currently joined JamSession. See [JamSession State](#jamsession-state)	|
| ``voting_type`` 	| string *optional*    	| Voting type of the User's currently joined JamSession. See [Avaliable Voting Types](#available-voting-types)	|

```json
{
    "name": "Joes Birthday Party",
    "active": true,
    "voting_type": "ip_voting"
}
```

***Response Body (JSON):***

| key      	        | value type        	| value description                                                                                     |
|----------	        |-------------------	|-------------------------------------------------------------------------------------------------------|
| ``label``         | string  	            | JamLabel of the User's currently joined JamSession                                       	            |
| ``name`` 	        | string             	| Name of the User's currently joined JamSession    	                                                |
| ``active`` 	    | boolean            	| State of the User's currently joined JamSession. See [JamSession State](#jamsession-state)	|
| ``voting_type`` 	| string             	| Voting type of the User's currently joined JamSession. See [Avaliable Voting Types](#available-voting-types)	|

```json
{
    "label": "KWXBZ",
    "name": "Joe's Birthday Party",
    "active": true,
    "voting_type": "ip_voting"
}
```

### Queue

#### 1. Add a Collection to the Queue of the User's joined JamSession

***Description***

Add a Spotify Collection (Playlist or Album) the the Current Queue of the User's joined JamSession. This can be used to add fallback music to a JamSession. The Songs in the Collection are voted in the queue with a virtual Vote. The User adding the Collection can still Vote for the added Songs. Note that the virtual Votes for the songs in the playlist is added with a date in the future. Therefore a real Vote of a User overrules the virtual Vote. For more details see [How voting works](#how-voting-works)

Requires the Uses to be the Host of a JamSession.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/queue/collection
```

***Request Body (JSON):***

| key      	        | value type            | value description                                              |
|----------	        |-------------------	|---------------------------------------------------             |
| ``collection`` 	| string *required*   	| Spotify ID of the collection.  |
| ``type`` 	        | string *required*   	| Type of he collection. <br> ``playlist`` for a playlist <br> ``album`` for an album  |

```json
{
	"collection": "3AGOiaoRXMSjswCLtuNqv5",
	"type": "album"
}
```

***Response Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``queue`` 	| array             	| Array of the Songs in the current Queue. See [Queue Song](#queue-song). The Queue Song Objects are personalized to the User requesting the Queue. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

#### 2. Delete a Song in the Queue of the User's joined JamSession


***Description***

Delete a Song in the Current Queue of the User's joined JamSession. Only songs wich are currently in the Queue can be deleted. This deletes all existing Votes for the Song but does not prevent a new Vote for the same Song.

Requires the User to be Host of a JamSession.

***Endpoint:***

```bash
Method: DELETE
URL: jamfactory.app/api/v1/queue/delete
```

***Request Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``track`` 	| string *required*   	| Spotify ID of the track which should be deleted. See [Spotify Track Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#track-object-full)  |

```json
{
	"track": "2374M0fQpWi3dLnB54qaLX"
}
```

***Response Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``queue`` 	| array             	| Array of the Songs in the current Queue. See [Queue Song](#queue-song). The Queue Song Objects are personalized to the User requesting the Queue. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

#### 3. Get the Queue of the User's joined JamSession


***Description***

Returns the Queue of the Users's joined JamSession. Don't query this endpoint on a regular basis. To get updates on chages of the queue use the provided Socket. See [Socket Reference](#socket-reference). Requires the User to be joined a JamSession.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/queue
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``queue`` 	| array             	| Array of the Songs in the current Queue. See [Queue Song](#queue-song). The Queue Song Objects are personalized to the User requesting the Queue. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

#### 4. Vote for a Song in the Queue of the User's joined JamSession


***Description***

Add a Vote from the User to a Song in the User's joined JamSession. If the Song does not exists in the queue, it will be added with the Users inital vote. If the User already Voted for the Song, the User's Vote will be removed. If a Song reaches zero Votes, the Song will be removed from the Queue. See [How voting works](#how-voting-works) for a more detailed description on how voting works.

Requires the User to be joined a JamSession.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/queue/vote
```

***Request Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``track`` 	| string *required*   	| Spotify ID of the track. See [Spotify Track Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#track-object-full)  |

```json
{
	"track": "2374M0fQpWi3dLnB54qaLX"
}
```

***Response Body (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``queue`` 	| array             	| Array of the Songs in the current Queue. See [Queue Song](#queue-song). The Queue Song Objects are personalized to the User requesting the Queue. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

### Spotify

#### 1. Get the User's Available Spotify Playback Devices


***Description***

Get information about the current user’s available devices. Requires the current user to be a JamSession Host.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/spotify/devices
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	    | value type        	| value description                                                                                          	|
|----------	    |-------------------	|------------------------------------------------------------------------------------------------------------	|
| ``devices``   | array 	            | Array of the available Spotify Playback Devices of the User. See [Spotify Device Object](https://developer.spotify.com/documentation/web-api/reference/player/get-a-users-available-devices/#device-object)                                    	|

```json
{
    "devices": "[]<Spotify Device Object>"
}
```

#### 2. Get the User's Available Spotify Playlists


***Description***

Get a list of the Spotify playlists owned or followed by the current user. Requires the current user to be a JamSession Host.

***Endpoint:***

```bash
Method: GET
URL: jamfactory.app/api/v1/spotify/playlists
```

***Request Body (Empty):***

***Response Body (JSON):***

| key      	        | value type        	| value description                                                                                          	|
|----------	        |-------------------	|------------------------------------------------------------------------------------------------------------	|
| ``playlists`` 	| [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Array of the Users avaliable Spotify Playlists as [Spotify Simplified Playlist Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#playlist-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object)

```json
{
    "playlists": "<Spotify Paging Object>"
}
```

#### 3. Search for an Item on Spotify


***Description***

Get Spotify Catalog information about playlists and tracks that match a keyword string. Requires the current user to be joined a JamSession.

***Endpoint:***

```bash
Method: PUT
URL: jamfactory.app/api/v1/spotify/search
```

***Request Body (JSON):***

| key      	| value type        	| value description                                                                                          	|
|----------	|-------------------	|------------------------------------------------------------------------------------------------------------	|
| ``text`` 	| string *required* 	| Search text. All search texts are completed with a ``*`` for autofill.                                       	|
| ``type`` 	| string *required* 	| Type of the searched item. Available:<br>``track`` for spotify track,<br>``playlist`` for spotify playlist 	|

```json
{
	"text": "abba",
	"type": "track"
}
```

***Response Body (JSON):***

| key      	        | value type        	| value description                                                                                          	|
|----------	        |-------------------	|------------------------------------------------------------------------------------------------------------	|
| ``artists`` 	    | [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Spotify Artists found with the submitted search as [Spotify Simplified Artists Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#artist-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) |
| ``albums`` 	    | [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Spotify Albums found with the submitted search as [Spotify Simplified Album Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#album-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) |
| ``playlists`` 	| [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Spotify Playlists found with the submitted search as [Spotify Simplified Playlist Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#playlist-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) |
| ``tracks`` 	    | [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) 	| Spotify Tracks found with the submitted search as [Spotify Simplified Track Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#track-object-simplified) wrapped in a [Spotify Paging Object](https://developer.spotify.com/documentation/web-api/reference/object-model/#paging-object) |

```json
{
    "artists": "<Spotify Paging Object>",
    "albums": "<Spotify Paging Object>",
    "playlists": "<Spotify Paging Object>",
    "tracks": "<Spotify Paging Object>"
}
```

# Socket Reference
JamFactory supports [Socket.IO](https://socket.io/) Websockets to update the Users information at certain events.
**Currently only Socket.IO Version <= 1.4 is supported**
When the a User has joined a JamSession as a Host or as a Guest, he can connect to the Socket.IO room created by the JamSession.
The connection will automatically join the right room based on the Session Cookie. 
The client only needs to open or close the connection and listen for the events.

***Example:***

Connect to the Socket
```js
    import * as io from 'socket.io-client';
    this.socket = io.connect('http://jamfactory.app:3000');
```

Listen to events
```js
     this.socket.on('<EventName>', (message: any) => {
          // <Code to handle the message>
     });
```

Close the connection if the User Leaves
```js
     this.socket.close();
```

## Socket Events

### Event: ``queue`` 

The Queue of the Users joined JamSession has changed.

***Message (JSON):***

| key      	    | value type            | value description                                              |
|----------	    |-------------------	|---------------------------------------------------             |
| ``queue`` 	| array             	| Array of the Songs in the current Queue. See [Queue Song](#queue-song). The Queue Song Objects are **not** personalized to the User requesting the Queue. |

```json
{
	"queue": "[]<Queue Song Object>"
}
```

### Event: ``playback`` 

Update on the current Playback State of the JamSession. This event is triggeredapproximately every second.

***Message (JSON):***

| key      	        | value type        	                                                                                            | value description                                     |
|----------	        |-------------------	                                                                                            |---------------------------------------------------    |
| ``playback``      | [Spotify Playback Object](https://developer.spotify.com/documentation/web-api/reference-beta/#endpoint-get-information-about-the-users-current-playback)  	| Playback state                                     	|
| ``device_id`` 	| string             	                                                                                            | Device ID of the currently selected playback device   |

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
| ``host``          | The Host closed the JamSession
| ``inactive`` 	    | The JamSession was closed due to inactivity 

---
[Back to top](#jamfactory)
> API Documentation Made with &#9829; by [thedevsaddam](https://github.com/thedevsaddam) | Generated at: 2020-10-03 13:28:49 by [docgen](https://github.com/thedevsaddam/docgen)
