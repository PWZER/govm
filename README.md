# GoVM

Golang multiple version manager

## Installation

```bash
go install github.com/PWZER/govm@latest
```

Or download the binary from [releases](https://github.com/PWZER/govm/releases/latest)

```bash
# example
wget https://github.com/PWZER/govm/releases/download/v0.1.0/govm-linux-amd64 -O govm

# make it executable
chmod +x govm
```

and set `PATH` and `GOPATH` environment variables

```bash
# govm will be linked current version gopath directory to ${HOME}/.govm/go
export GOPATH=${HOME}/.govm/go

# add govm binary install path and ${GOPATH}/bin to PATH
export PATH=${PATH}:${HOME}/.local/bin:${HOME}/.govm/go/bin
```

## Usage

```bash
# display govm info
govm

# list remote versions
govm ls --remote

# list local versions
govm ls

# install a version
govm install go1.23.0

# install with proxy, support environment variable HTTP_PROXY, HTTPS_PROXY, NO_PROXY
HTTP_PROXY=http://proxy:port govm install go1.23.0

# install specify mirror
govm install go1.23.0 --mirror https://golang.google.cn/dl/

# use or change the go version
govm use go1.23.0
```
