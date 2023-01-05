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
