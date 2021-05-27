# Jamfactory Backend
<p align="center">
    <img src="docs/logo.svg" alt="Logo" width="80" height="80">
</p>

<p align="center">
Checkout the live version on
<a href="https://jamfactory.app"><strong>Jamfactory.app</strong></a>
</p>


## What is JamFactory

JamFactory is a collaborative playback controller. A JamSession is a private party with **one** host to set it up, and many attendees to join in.

Your friends or party guests can vote for the songs they like and want to listen to, and the song with the most votes gets played next.
JamFactory acts as the conductor of your music and is not the playback device itself. Therefore, you are free to **choose your own playback device**.

The host of a JamSession has to have a Spotify premium account, the guests can join the without a Spotify account.

JamFactory consists of independent applications that form the ecosystem together. This project contains the backend that provides the necessary api to create, join and control a JamSession.
It also acts as the Conductor that controls the playback and communicates to the Spotify api.

### Built With

JamFactory is build among others using these awesome projects 
* [go](https://golang.org/)
* [zmb3/spotify](https://github.com/zmb3/spotify)
* [gorilla ](https://www.gorillatoolkit.org/)
* [Redis](https://redis.io/)

### Last Release

``V0.1.0`` [Release Notes](./RELEASE.md)

## Getting started

To understand how the JamFactory backend works, which API Endpoints are available and how it can be used to create a JamSession read the [Documentation](./docs/documentation.md) 


### Installation

The JamFactory backend can either be installed using docker (*recommended*) or build and installed locally.
Helm charts for kubernetes deployments are available at [JamFactoryApp/jamfactory-helm](https://github.com/JamFactoryApp/jamfactory-helm).

#### Initial setup

* Clone the repository to your desired location
  ```sh
  git clone https://github.com/JamFactoryApp/jamfactory-backend.git
  ```
* Create a Spotify App on the [Developer Dashboard](https://developer.spotify.com/dashboard)

* Create an ``.env`` file and fill out the information. See [.env.example](./.env.example) for an example ``.env`` file. It is recommended to use a very long password for the redis db.

#### Docker installation

* Create a users.acl file in the ``/redis`` folder with. Make sure you use the same password as in the .env file. See [.env.example](./redis/users.acl.example) for an example ``users.acl`` file.

* Create and start the docker containers using ``docker-compose up -d``

#### Local installation
TODO



