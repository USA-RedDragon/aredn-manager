#!/bin/sh

# This is Ubuntu 22.04 LTS (Jammy)

# Install Docker and Git
apt update
apt install -y docker.io git

systemctl enable --now docker

# Add the ubuntu user to the Docker group
usermod -aG docker ubuntu

# Clone this repo
git clone https://github.com/USA-RedDragon/aredn-virtual-node -b main

# Build the Docker image
cd aredn-virtual-node
docker build -t aredn-virtual-node .

# Run the Docker image
docker run \
    --cap-add=NET_ADMIN \
    --privileged \
    -e CONFIGURATION_JSON='${configuration_json}' \
    -e SERVER_NAME=${server_name} \
    --device /dev/net/tun \
    --name ${server_name} \
    -p 5525:5525 \
    -d \
    --restart unless-stopped \
    aredn-virtual-node
