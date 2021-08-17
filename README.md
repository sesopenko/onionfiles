# README.md

## Requirements

* Tested in ubuntu 20.04
* `sudo apt-get install build-essential`

## Build

```bash
# download dependencies
make get
# build executable
make build
```

## Running

```bash
docker run --read-noly -v /path/to/files:/app/static/files/ sesopenko/onionfiles
```