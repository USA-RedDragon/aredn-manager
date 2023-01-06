#!/bin/sh

# This is Ubuntu 22.04 LTS (Jammy)

apt update
apt upgrade -y
apt install -y docker.io

systemctl enable --now docker
systemctl disable --now snapd.service
systemctl disable --now snap.amazon-ssm-agent.amazon-ssm-agent.service

echo 'wireguard' >> /etc/modules-load.d/modules.conf
modprobe wireguard

# Add the ubuntu user to the Docker group
usermod -aG docker ubuntu

# Clone this repo
docker pull ghcr.io/usa-reddragon/aredn-virtual-node:main

# Run the Docker image
docker run \
    --cap-add=NET_ADMIN \
    --privileged \
    -e CONFIGURATION_JSON='${configuration_json}' \
    -e SERVER_NAME=${server_name} \
    -e WIREGUARD_TAP_ADDRESS=${wireguard_tap_address} \
    -e NUM_WIREGUARD_PEERS=${num_wireguard_peers} \
    --device /dev/net/tun \
    --name ${server_name} \
    -p 5525:5525 \
    -p 51820:51820/udp \
    -d \
    --restart unless-stopped \
    ghcr.io/usa-reddragon/aredn-virtual-node:main
