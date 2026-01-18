FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

# Install system dependencies
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    build-essential \
    git \
    curl \
    wget \
    ca-certificates \
    wireguard-tools \
    iptables \
    iproute2 \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*

# Set up non-root user
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# Install make
RUN apt-get update && apt-get install -y make && rm -rf /var/lib/apt/lists/*
