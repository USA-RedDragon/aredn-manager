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

docker network create --subnet=10.54.25.0/24 aredn-net

LOGGING="--log-driver=awslogs --log-opt awslogs-region=${region} --log-opt awslogs-group=${awslogs-group} --log-opt awslogs-create-group=true"

mkdir -p /docker-data

# Try to mount /dev/sdf first, then if it fails, format it
if ! mount -t ext4 /dev/nvme1n1 /docker-data; then
    mkfs.ext4 /dev/nvme1n1
    mount -t ext4 /dev/nvme1n1 /docker-data
fi

mkdir -p /docker-data/netdata
chown -R root:201 /docker-data/netdata
chmod -R g+w /docker-data/netdata

CH="http://${server_name}.local.mesh:81${extra_cors_hosts}"

# If ${supernode} is set
if [[ "${supernode}" != "" ]]; then
    # Add http://${server_name}.${supernode_zone}.mesh:81 to the CH
    CH="$CH,http://${server_name}.${supernode_zone}.mesh:81"
fi

# Run the Docker image
docker run \
    --cap-add=NET_ADMIN \
    --privileged \
    -e MAP_CONFIG='${map_config_json}' \
    -e PG_HOST='${pg_host}' \
    -e PG_USER='${pg_user}' \
    -e PG_PASSWORD='${pg_pass}' \
    -e PG_DATABASE='${pg_db}' \
    -e SESSION_SECRET='${session_secret}' \
    -e PASSWORD_SALT='${password_salt}' \
    -e CORS_HOSTS="$CH" \
    -e INIT_ADMIN_USER_PASSWORD='${init_admin_user_password}' \
    -e SERVER_NAME=${server_name} \
    -e SUPERNODE="${supernode}" \
    -e SERVER_LON='${server_lon}' \
    -e SERVER_LAT='${server_lat}' \
    -e SERVER_GRIDSQUARE=${server_gridsquare} \
    -e WIREGUARD_TAP_ADDRESS=${wireguard_tap_address} \
    -e WIREGUARD_PEER_PUBLICKEY=${wireguard_peer_publickey} \
    -e WIREGUARD_SERVER_PRIVATEKEY=${wireguard_server_privatekey} \
    -e NODE_IP=${node_ip} \
    -e SUPERNODE_ZONE=${supernode_zone} \
    --device /dev/net/tun \
    --name ${server_name} \
    -p 5525:5525 \
    -p 51820:51820/udp \
    -d \
    --restart unless-stopped \
    $LOGGING \
    --net aredn-net --ip ${node_ip} \
    ghcr.io/usa-reddragon/aredn-virtual-node:main

docker run \
    -d \
    --name watchtower \
    $LOGGING \
    -v /var/run/docker.sock:/var/run/docker.sock \
    --restart=unless-stopped \
    containrrr/watchtower \
    --cleanup

docker run \
    --network=container:${server_name} \
    -v /etc/passwd:/host/etc/passwd:ro \
    -v /etc/group:/host/etc/group:ro \
    -v /proc:/host/proc:ro \
    -v /sys:/host/sys:ro \
    -v /etc/os-release:/host/etc/os-release:ro \
    --restart unless-stopped \
    --cap-add SYS_PTRACE \
    --security-opt apparmor=unconfined \
    -d \
    -v /docker-data/netdata/etc:/etc/netdata \
    -v /docker-data/netdata/var:/var/lib/netdata \
    -v /docker-data/netdata/cache:/var/cache/netdata \
    $LOGGING \
    --name netdata \
    netdata/netdata
