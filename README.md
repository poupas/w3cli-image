# w3cli-image

This repository hosts the files needed to build the Docker image for the new [web3.storage CLI](https://web3.storage/docs/w3cli/). It will be hosted here on Docker Hub: [https://hub.docker.com/r/rocketpool/w3cli](https://hub.docker.com/r/rocketpool/w3cli)

The image is based off of the [slim Debian 12 image with NodeJS 20 support](https://github.com/nodejs/docker-node/blob/dbc174542d51f03535f6513391f569e3b93a91dd/20/bookworm-slim/Dockerfile), tagged as `node:lts-bookworm-slim` on Docker Hub. It adds a new user named `w3s` with a home directory set to `/w3state` (which is where all of `w3cli`'s keys and state are saved), installs web3.storage's CLI, and adds a simple relayer that can invoke `w3` commands remotely via a Unix domain socket. The relayer is intended to be used by other Docker containers that need access to a few `w3` features, such as `up`.


## The Relayer

The relayer is a simple application that acts as an HTTP server over a Unix domain socket and relays a select few commands to / responses from `w3`. It's primarily intended to be used by the Rocket Pool watchtower so rewards files can be uploaded to web3.socket since the Go library has been deprecated. The binary for the relayer is located at `/usr/local/bin/w3cli_relayer`.

To run it, use the following syntax:
```
w3cli_relayer -s <path to create the socket file> -r <path to the rewards_trees folder>
```
where:
- `-s` should be the path (inside this container) to the mountpoint of a named Docker volume shared exclusively between the watchtower container and this container
- `-r` should be the path (inside this container) to the mountpoint of the `rewards_trees` folder that the watchtower saves its generated rewards artifacts to

The relayer listens on `http://w3cli` and supports the following `GET` routes:
- `/whoami`: runs `w3 whoami`
- `/login?email=<email>`: runs `w3 login <email>`
- `/space-create`: runs `w3 space create rp_odao --no-recovery`, which is just a simple default space for the Rocket Pool watchtower in case the user doesn't want to deal with making a space on their own
- `/up?file=<file>`: runs `w3 up <file> --no-wrap --json`, where `<file>` must be the filename of a file in the directory provided with the `-r` argument.

If you don't need to use the relayer and just want to use `w3` via a container, you can ignore this aspect of the image.


## How to Use the Image


### Invoking `w3` From the CLI

The entrypoint for the image is simply the `w3` command (the web3.storage CLI). It runs as the `w3s` user.

Prior to running, you will need a volume you want to capture the CLI's state and private keys. This is required for your login state and keys to persist properly when the container is stopped and its filesystem is reset. The CLI stores all of its state in the `~/.config` directory, so the volume will map to `/w3state/.config`.


In this example, we'll use a Docker volume named `w3cli_cfg`.

To run the CLI via the image, for example with a `whoami` command, run a `docker run` command with the volume mounted and the arguments to `w3` specified at the end:

```
docker run --rm -v w3cli_cfg:/w3state/.config rocketpool/w3cli whoami
```

Replace `whoami` with whatever arguments you want to pass to `w3`.


### Running the Relayer

To run the relayer, which will make the image persistent until shut down, you can invoke it as in the following example:
```
mkdir /tmp/w3cli_socket && sudo chmod 777 /tmp/w3cli_socket

docker run --rm -it --entrypoint w3cli_relayer -v w3cli_cfg:/w3state/.config -v /home/user/.rocketpool/data/rewards-trees:/rewards-trees -v /tmp/w3cli_socket:/socket rocketpool/w3cli -s /socket/w3cli_relayer.sock -r /rewards-trees
```

Note that whatever folder you use as the socket volume will either need `0777` permissions (so the relayer can create the socket) or be `chown`'d to the `w3s` user's UID / GID (`65530` by default).

Once running, you can test the connection with a simple cURL command:

```
sudo curl -X GET --unix-socket /tmp/w3cli_socket/w3cli_relayer.sock http://w3cli/whoami
```
(here running as `sudo` so it can write to the socket file).
If successful, you'll see some log output on the tab running the Docker image and get a response back stating the user's DID from `w3`.


## Building the Image

To build the image and push it to Docker Hub, use the included `build-release.sh` script:

```
./build-release.sh -av v6.1.0-1
```

where the version number after the `-v` flag should be the version of the included web3.storage CLI, followed by a hyphen, then the iteration of the image built against that version (in case multiple need to be made).