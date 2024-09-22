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

## Usage

```bash
# display info
govm

# list remote versions
govm ls --remote

# list local versions
govm ls

# install a version
govm install go1.23.0

# use a version
govm use go1.23.0
```
