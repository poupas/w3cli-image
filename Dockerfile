# This builds the CLI relayer
FROM golang:1.21-bookworm AS builder
COPY ./relayer /relayer
RUN cd /relayer && go build


# This image wraps the web3.storage CLI
FROM node:lts-bookworm-slim

ARG USER=w3s
ARG UID=65530
ARG GID=65530

RUN apt update && \
	# Install git, which the web3.storage CLI requires
	apt install git -y && \
	apt-get clean && \
	rm -rf /var/lib/apt/lists/* && \
	# Install the web3.storage CLI
	npm install -g @web3-storage/w3cli && \
	# Make a new user
	groupadd -g "${GID}" "${USER}" && useradd -m -d /w3state -u "${UID}" -g "${GID}" "${USER}"

# Copy the relayer
COPY --from=builder /relayer/w3cli_relayer /usr/local/bin/w3cli_relayer

# Run w3 by default as the specified user
USER $USER
WORKDIR $HOME
ENTRYPOINT ["w3"]
