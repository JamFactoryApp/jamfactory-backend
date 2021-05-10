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

Jamfactory consists of independent applications that form the ecosystem together. This project contains the backend that provides the necessary api to create, join and control a JamSession. See [API Documentation](./docs/documentation.md)
It also acts as the Conductor that controls the playback and communicates to the Spotify api.

### Built With

JamFactory is build among others using these awesome projects 
* [go](https://golang.org/)
* [zmb3/spotify](https://github.com/zmb3/spotify)
* [gorilla ](https://www.gorillatoolkit.org/)
* [Redis](https://redis.io/)

### Last Release

[Release Notes](./RELEASE.md)

## Installation

Jamfactory backend can either be installed using docker (*recommended*) or build and installed locally

## Initial setup

* Clone the repository to your desired location
  ```sh
  git clone https://github.com/JamFactoryApp/jamfactory-backend.git
  ```
* Create a Spotify App 

* Create an ``.env`` file. See [.env.example](./.env.example) for an example ``.env`` file.


### Docker installation

TODO

### Local installation

TODO



