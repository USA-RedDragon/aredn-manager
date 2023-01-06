#/bin/sh

# Trap signals and exit
trap "exit 0" SIGHUP SIGINT SIGTERM

/usr/bin/blockknownencryption

# If CONFIGURATION_JSON is not set
if [ -z "$CONFIGURATION_JSON" ]; then
    echo "No configuration JSON provided, exiting"
    exit 1
fi

if [ -z "$SERVER_NAME" ]; then
    echo "No server name provided, exiting"
    exit 1
fi

# If NUM_WIREGUARD_PEERS is set and greater than 0
if [ -n "$NUM_WIREGUARD_PEERS" ] && [ "$NUM_WIREGUARD_PEERS" -gt 0 ]; then
    ip link add dev wg0 type wireguard
    ip address add dev wg0 ${WIREGUARD_TAP_ADDRESS}/32
    if [ -z "$WIREGUARD_TAP_ADDRESS" ]; then
        echo "No WireGuard address provided, exiting"
        exit 1
    fi

    mkdir -p /etc/wireguard/keys

    wg genkey | tee /etc/wireguard/keys/server.key | wg pubkey > /etc/wireguard/keys/server.pub

    # For each peer, create a private key and save it into /etc/wireguard/keys/peer-<peer number>.key
    export WIREGUARD_TMP_ADDRESS=$WIREGUARD_TAP_ADDRESS
    for i in $(seq 1 $NUM_WIREGUARD_PEERS); do
        wg genkey | tee /etc/wireguard/keys/peer-$i.key | wg pubkey > /etc/wireguard/keys/peer-$i.pub
        export WG_TAP_PLUS_I=$(echo $WIREGUARD_TMP_ADDRESS | awk -F. '{print $1"."$2"."$3"."$4+1}')
        export WIREGUARD_TMP_ADDRESS=$WG_TAP_PLUS_I
        wg set wg0 peer $(cat /etc/wireguard/keys/peer-$i.pub) allowed-ips $WG_TAP_PLUS_I/32,10.0.0.0/8

        export PRIVATE_KEY=$(cat /etc/wireguard/keys/peer-$i.key)
        export PEER_ADDRESS=$WG_TAP_PLUS_I
        export PUBLIC_KEY=$(cat /etc/wireguard/keys/server.pub)
        envsubst < /tpl/wireguard-peer.conf > /etc/wireguard/peer-$i.conf
    done

    chmod 400 /etc/wireguard/keys/*
    chmod 400 /etc/wireguard/*.conf

    wg set wg0 listen-port 51820 private-key /etc/wireguard/keys/server.key

    # Cross-VPN traffic OK
    iptables -A FORWARD -i wg0 -o wg0 -j ACCEPT
    # No internet access for the VPN clients
    iptables -A FORWARD -i wg0 -o eth0 -j REJECT
    iptables -A FORWARD -i eth0 -o wg0 -j REJECT

    iptables -t mangle -A PREROUTING -i wg0 -j MARK --set-mark 0x30
    iptables -t nat -A POSTROUTING ! -o wg0 -m mark --mark 0x30 -j MASQUERADE
fi

# Here is an exmaple of a configuration JSON
# {
#     "clients": [
#         { "name": "KI5VMF-MAIN", "net": "172.31.180.16", "pwd": "changeme"},
#         { "name": "KI5VMF-SECOND", "net": "172.31.180.20", "pwd": "changemetoo"}
#     ]
# }

CLIENTS=$(echo $CONFIGURATION_JSON | jq -c '.clients[]')
TUN=50
CLIENT_CONFIGS=""

for CLIENT in $CLIENTS; do
    export NAME=$(echo $CLIENT | jq -r '.name')
    export NET=$(echo $CLIENT | jq -r '.net')
    export PWD=$(echo $CLIENT | jq -r '.pwd')
    export DASHED_NET=$(echo $NET | sed 's/\./-/g')
    export IP_PLUS_1=$(echo $NET | awk -F. '{print $1"."$2"."$3"."$4+1}')
    export IP_PLUS_2=$(echo $NET | awk -F. '{print $1"."$2"."$3"."$4+2}')
    export TUN=$TUN

    LATEST_CONFIG="$(envsubst < /tpl/client.conf)"
    export CLIENT_CONFIGS=$(echo -e "$CLIENT_CONFIGS\n\n$LATEST_CONFIG")

    if [ -n "$NUM_WIREGUARD_PEERS" ] && [ "$NUM_WIREGUARD_PEERS" -gt 0 ]; then
        # Allowing all active and related connections
        iptables -A FORWARD -i wg0 -o tun$TUN -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
        iptables -A FORWARD -i tun$TUN -o wg0 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

        # Cross-talk between tun and wg0
        iptables -A FORWARD -i wg0 -o tun$TUN -j ACCEPT
        iptables -A FORWARD -i tun$TUN -o wg0 -j ACCEPT

        ip link set wg0 up
    fi

    # No internet access for the tunnels
    iptables -A FORWARD -i tun$TUN -o eth0 -j REJECT
    iptables -A FORWARD -i eth0 -o tun$TUN -j REJECT

    # Increment the TUN number
    TUN=$((TUN+1))
done
envsubst < /tpl/vtundsrv.conf > /etc/vtundsrv.conf

export SERVER_NAME=$SERVER_NAME
export IFACES=$(seq 50 $TUN | xargs -I{} echo -n "\"tun{}\" ")
export TUNNELS=$(envsubst < /tpl/olsrd-tunnel.conf)

mkdir -p /etc/olsrd/
envsubst < /tpl/olsrd.conf > /etc/olsrd/olsrd.conf

rsyslogd

cat <<EOF > /tmp/resolv.conf.auto
nameserver 1.1.1.1
nameserver 1.0.0.1
EOF

echo -e 'search local.mesh\nnameserver 127.0.0.1' > /etc/resolv.conf

dnsmasq

vtund -s -f /etc/vtundsrv.conf

olsrd

tail -f /var/log/messages
