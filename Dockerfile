# This image wraps the web3.storage CLI
FROM node:lts-bookworm-slim

ARG USER=w3s
ARG UID=65530
ARG GID=65530
ENV USER=$USER
ENV UID=$UID
ENV GID=$GID

RUN apt update && \
	# Install git, which the web3.storage CLI requires
	apt install git -y && \
	apt-get clean && \
	rm -rf /var/lib/apt/lists/* && \
	# Install the web3.storage CLI
	npm install -g @web3-storage/w3cli && \
	# Make a new user
	groupadd -g "${GID}" "${USER}" && useradd -m -u "${UID}" -g "${GID}" "${USER}"

USER $USER
WORKDIR /home/$USER
ENTRYPOINT ["w3"]
