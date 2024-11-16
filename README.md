<p align="center">
  <img src="https://github.com/coolapso/go-live-server/blob/main/test/images/goliveserver.png" width="200" >
</p>

# go-live-server
[![Release](https://github.com/coolapso/go-live-server/actions/workflows/release.yaml/badge.svg?branch=main)](https://github.com/coolapso/go-live-server/actions/workflows/release.yaml)
[![Docker image version](https://img.shields.io/docker/v/coolapso/go-live-server/latest?logo=docker)](https://hub.docker.com/r/coolapso/go-live-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/coolapso/go-live-server)](https://goreportcard.com/report/github.com/coolapso/go-live-server)
![GitHub Sponsors](https://img.shields.io/github/sponsors/coolapso?style=flat&logo=githubsponsors)

Go implementation of [tapio/liveserver](https://github.com/tapio/live-server). A Simple development web server written in go with live reload capabilities

## How it works

The server functions similarly to [tapio/liveserver](https://github.com/tapio/live-server). It serves a directory and its subdirectories, and establishes a WebSocket connection by injecting a JavaScript snippet into every HTML file it serves. The server monitors changes in the directories it serves, and when changes occur, it sends a message via WebSocket to instruct the browser to reload the page.

## Installation 

### Docker

You can run go-live-server with docker: 

```
docker run -v $(pwd):/data -p 8080:8080 coolapso/go-live-server:latest
```

Now you can edit your html files and let go-live-server work its magic from within the docker container

### Go Install

#### Latest version 

`go install github.com/coolapso/go-live-server`

#### Specific version

`go install github.com/coolapso/go-live-server@v1.0.0`

### AUR

On Arch linux you can use the AUR `go-live-server-bin`

### Linux Script

It is also possible to install on any linux distro with the installation script

#### Latest version

```
curl -L http://live-server.coolapso.sh/install.sh | bash
```

#### Specific version

```
curl -L http://live-server.coolapso.sh/install.sh | VERSION="v1.1.0" bash
```

### Manual install

* Grab the binary from the [releases page](https://github.com/coolapso/go-live-server/releases).
* Extract the binary
* Execute it

## Usage 

```
go-live-server is a simple development webserver with live reloading capabilityes.
It allows you to quickly edit and visualize changes when developing simple html and css files

Usage:
  live-server [flags]
  live-server [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print live-server version

Flags:
      --browser            Enable or disable automatic opening of the browser (default true)
  -h, --help               help for live-server
      --open-file string   Specify the relative path to open the browser in the directory being served
  -p, --port string        The port server is going to listen on (default ":8080")
  -d, --watch-dir string   Sets the directory to watch for (default "./")

Use "live-server [command] --help" for more information about a command.
```

## Build 

### With makefile

`make build`

### Manually

`go build -o go-live-server`

# Contributions

Improvements and suggestions are always welcome, feel free to check for any open issues, open a new Issue or Pull Request

If you like this project and want to support / contribute in a different way you can always [:heart: Sponsor Me](https://github.com/sponsors/coolapso) or

<a href="https://www.buymeacoffee.com/coolapso" target="_blank">
  <img src="https://cdn.buymeacoffee.com/buttons/default-yellow.png" alt="Buy Me A Coffee" style="height: 51px !important;width: 217px !important;" />
</a>



Also consider supporting [tapio/live-server](https://github.com/tapio/live-server) which inspired this project
