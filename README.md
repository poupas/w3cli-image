# w3cli-image

This repository hosts the files needed to build the Docker image for the new [web3.storage CLI](https://web3.storage/docs/w3cli/). It will be hosted here on Docker Hub: [https://hub.docker.com/r/rocketpool/w3cli](https://hub.docker.com/r/rocketpool/w3cli)

The image is based off of the [slim Debian 12 image with NodeJS 20 support](https://github.com/nodejs/docker-node/blob/dbc174542d51f03535f6513391f569e3b93a91dd/20/bookworm-slim/Dockerfile), tagged as `node:lts-bookworm-slim` on Docker Hub. It adds a new user named `w3s` and installed web3.storage's CLI, but is otherwise unmodified.


## How to Use the Image

The entrypoint for the image is simply the `w3` command (the web3.storage CLI). It runs as the `w3s` user.

Prior to running, you will need a volume you want to capture the CLI's state and private keys. This is required for your login state and keys to persist properly when the container is stopped and its filesystem is reset. The CLI stores all of its state in the `~/.config` directory, so the volume will map to `/home/w3s/.config`.


In this example, we'll use a Docker volume named `w3cli_cfg`.

To run the CLI via the image, for example with a `whoami` command, run a `docker run` command with the volume mounted and the arguments to `w3` specified at the end:

```
docker run --rm -v w3cli_cfg:/home/w3s/.config rocketpool/w3cli:latest whoami
```

Replace `whoami` with whatever arguments you want to pass to `w3`.


## Building the Image

To build the image and push it to Docker Hub, use the included `build-release.sh` script:

```
./build-release.sh -av v6.1.0-1
```

where the version number after the `-v` flag should be the version of the included web3.storage CLI, followed by a hyphen, then the iteration of the image built against that version (in case multiple need to be made).