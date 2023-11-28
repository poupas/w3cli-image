#!/bin/bash

# Build the images
build_docker() {
    echo -n "Building Docker images... "
    docker buildx build --platform=linux/amd64 -t rocketpool/w3cli:$VERSION-amd64 --load .
    docker buildx build --platform=linux/arm64 -t rocketpool/w3cli:$VERSION-arm64 --load .
    docker push rocketpool/w3cli:$VERSION-amd64
    docker push rocketpool/w3cli:$VERSION-arm64
    echo "done!"
}

# Build the manifest
build_manifest() {
    echo -n "Building Docker manifest... "
    rm -f ~/.docker/manifests/docker.io_rocketpool_w3cli-$VERSION
    rm -f ~/.docker/manifests/docker.io_rocketpool_w3cli-latest
    docker manifest create rocketpool/w3cli:$VERSION --amend rocketpool/w3cli:$VERSION-amd64 --amend rocketpool/w3cli:$VERSION-arm64
    docker manifest create rocketpool/w3cli:latest --amend rocketpool/w3cli:$VERSION-amd64 --amend rocketpool/w3cli:$VERSION-arm64
    echo "done!"
    echo -n "Pushing to Docker Hub... "
    docker manifest push --purge rocketpool/w3cli:$VERSION
    docker manifest push --purge rocketpool/w3cli:latest
    echo "done!"
}

# Print usage
usage() {
    echo "Usage: build-release.sh [options] -v <version number>"
    echo "This script builds the w3cli image and pushes it to Docker Hub."
    echo "Options:"
    echo $'\t-a\tBuild and push all of the artifacts'
    echo $'\t-d\tBuild the Docker containers and push them to Docker Hub'
    echo $'\t-m\tBuild the Docker manifests and push them to Docker Hub'
    exit 0
}

# =================
# === Main Body ===
# =================

# Get the version
while getopts "admv:" FLAG; do
    case "$FLAG" in
        a) DOCKER=true MANIFEST=true ;;
        d) DOCKER=true ;;
        m) MANIFEST=true ;;
        v) VERSION="$OPTARG" ;;
        *) usage ;;
    esac
done
if [ -z "$VERSION" ]; then
    usage
fi

# Build the artifacts
if [ "$DOCKER" = true ]; then
    build_docker
fi
if [ "$MANIFEST" = true ]; then
    build_manifest
fi
