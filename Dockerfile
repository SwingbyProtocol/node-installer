# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:1.14 as builder

# Add Maintainer Info
LABEL maintainer="Yusaku Senga <yusaku@swingby.network>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN make build-linux-amd64

##### Start a new stage from scratch #######
FROM ubuntu:18.04

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bin/bot_linux_amd64 /app/bot_linux_amd64
COPY playbooks /app/playbooks

# Install ansible
RUN apt-get update \
    && apt-get install -y --no-install-recommends software-properties-common \
    && add-apt-repository ppa:ansible/ansible \
    && apt-get install -y --no-install-recommends \
    ansible openssh-client \
    && apt-get -y clean \
    && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/app/bot_linux_amd64"]

