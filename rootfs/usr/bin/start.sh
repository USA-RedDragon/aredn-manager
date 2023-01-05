#/bin/sh

# Trap signals and exit
trap "exit 0" SIGHUP SIGINT SIGTERM

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
