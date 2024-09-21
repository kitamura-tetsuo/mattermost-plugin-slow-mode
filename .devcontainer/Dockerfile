# Use Node.js v16 base image (bookworm variant)
FROM node:16-bookworm

# Install necessary packages including sudo
RUN apt-get update && \
    apt-get install -y sudo curl less man-db

# Install Go version 1.21.13
RUN curl -LO https://go.dev/dl/go1.21.13.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz && \
    rm go1.21.13.linux-amd64.tar.gz

# Set environment variables for Go
ENV PATH="/usr/local/go/bin:${PATH}"

# Create a vscode user and allow passwordless sudo
RUN useradd -m vscode && \
    echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Upgrade npm to version 8
RUN npm install -g npm@8

# Create workspace directory and set permissions for vscode user
RUN mkdir -p /workspaces/mattermost-plugin-slow-mode && \
    chown -R vscode:vscode /workspaces/mattermost-plugin-slow-mode

# Set default user to vscode
USER vscode

# Set workspace directory
WORKDIR /workspaces/mattermost-plugin-slow-mode

# Set `DEVCONTAINER` environment variable to assist with devcontainer detection
ENV DEVCONTAINER=true
