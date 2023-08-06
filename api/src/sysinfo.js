const fs = require('fs');
const os = require('os');
const http = require('http');
const dns = require('dns').promises;

const obj = {
    lon: '-97.461861',
    lat: '35.324456',
    sysinfo: {
        uptime: os.uptime(),
        loads: [
            os.loadavg()[0],
            os.loadavg()[1],
            os.loadavg()[2],
        ],
    },
    api_version: '1.11',
    meshrf: {
        status: 'off',
    },
    grid_square: 'EM15gh',
    node: 'KI5VMF-CLOUD-TUNNEL',
    node_details: {
        description: 'AREDN Cloud Tunnel',
        model: 'Software',
        mesh_gateway: '1',
        board_id: '0x0000',
        firmware_mfg: 'USA-RedDragon',
        firmware_version: '1.0.0',
    },
    tunnels: {
        active_tunnel_count: getActiveTunnels(),
    },
    lqm: {
        enabled: false,
    },
    // Below items are to be filled in
    // os.networkInterfaces()
    interfaces: getInterfaces(),
    // hosts are parsed from /var/run/hosts_olsr
    hosts: getHosts(),
    // services are parsed from /var/run/services_olsr
    services: getServices(),
    // link info can come from http://localhost:9090/links
    link_info: getLinkInfo(),
};

function getInterfaces() {
    // Return an array of network interface objects with the following fields (if possible):
    // mac, name, ip

    const interfaces = [];

    for (const [key, value] of Object.entries(os.networkInterfaces())) {
        if (key === 'lo' || key.startsWith('wg')) {
            continue;
        }
        for (const v of value) {
            const iface = {
                name: key,
                ip: v.address,
            };
            if (v.internal === false && v.family === 'IPv4') {
                if (v.mac !== '00:00:00:00:00:00') {
                    iface.mac = v.mac;
                }
                interfaces.push(iface);
            }
        }
    }

    return interfaces;
}

function getHosts() {
    // This needs to be parsed from /var/run/hosts_olsr
    // The file can have comments prefixed with #
    // The file can have blank lines
    // The format of a line is the same as /etc/hosts
    // Return an array of host objects with the following fields:
    // ip, name

    const hosts = [];
    fs.readFileSync('/var/run/hosts_olsr', 'utf8').split(/\r?\n/).forEach(function (line) {
        if (line.startsWith('#') || line === '') {
            return;
        }
        const parts = line.split(/\s+/);
        const midRegex = /^mid\d+\./i;
        const dtdLinkRegex = /^dtdlink\./i;
        if (!(parts[1] == 'localhost') && !parts[1].match(midRegex) && !parts[1].match(dtdLinkRegex)) {
            const host = {
                ip: parts[0],
                name: parts[1],
            };
            hosts.push(host);
        }
    });
    return hosts;
}

function getServices() {
    // This needs to be parsed from /var/run/services_olsr
    // The file can have comments prefixed with #
    // The file can have blank lines
    // The format of a line is: http://N5AZQ-GL-INET-AR300M-0:0/|tcp|TeamTalk 5-N5AZQ   #10.33.212.177
    // Return an array of service objects with the following fields:
    // ip, name, protocol, link
    // Links with a port of 0 are non-http links, so the link should be ""

    const services = [];
    fs.readFileSync('/var/run/services_olsr', 'utf8').split(/\r?\n/).forEach(function (line) {
        if (line.startsWith('#') || line === '') {
            return;
        }
        // Use the following regex to split link, protocol, name, ip
        const regex = /^([^|]*)\|(.+)\|(.+)#(.+)/;
        const matches = line.match(regex);
        if (matches === null || matches.length !== 4) {
            return;
        }
        let link = matches[0];
        const protocol = matches[1];
        const name = matches[2];
        let ip = matches[3];
        // If the link ends with :0/, then it is a non-http link, so set link to ""
        if (link.endsWith(':0/')) {
            link = '';
        }

        // If the ip is " my own service", then get the IP of the main network interface
        if (ip === ' my own service') {
            ip = '127.0.0.1';
        }

        services.push({
            link,
            protocol,
            name,
            ip,
        });
    });
}

async function getLinkInfo() {
    // Return an object with fields matching ip addresses of links
    // We can download http://127.0.0.1:9090/links and parse the JSON. The `links` field is what we want.
    // The output link object should look contain
    // helloTime, lostLinkTime, linkQuality, vtime, linkCost, linkType, hostname, previousLinkStatus,
    // currentLinkStatus, neighborLinkQuality, symmetryTime, seqnoValid, pending, lossHelloInterval,
    // lossMultiplier, hysteresis, seqno, lossTime, validityTime, olsrInterface, lastHelloTime, asymmetryTime

    const linkInfo = {};
    const promise = new Promise((resolve, reject) => {
        http.get('http://localhost:9090/links', (resp) => {
            let data = '';
            resp.on('data', (chunk) => {
                data += chunk;
            });

            resp.on('end', () => {
                resolve(JSON.parse(data));
            });

            resp.on('error', (err) => {
                reject(err);
            });
        });
    });

    const data = await promise;

    if (data === null) {
        return linkInfo;
    } else if (data.links === undefined) {
        return linkInfo;
    }
    for (const link of data.links) {
        let hostname = '';
        // Do a reverse lookup on the IP address to get the hostname
        try {
            hostnameData = await dns.reverse(link.remoteIP);
            if (hostnameData.length > 0) {
                hostname = hostnameData[0];
                // Strip off mid\d. from the hostname if it exists
                let regex = /^mid\d\.(.+)/;
                let matches = hostname.match(regex);
                if (matches !== null && matches.length === 2) {
                    hostname = matches[1];
                }
                // Strip off dtdlink. from the hostname if it exists
                regex = /^dtdlink\.(.+)/;
                matches = hostname.match(regex);
                if (matches !== null && matches.length === 2) {
                    hostname = matches[1];
                }
            }
        } catch (err) {
            console.log(err);
        }

        const newObj = {
            helloTime: link.helloTime,
            lostLinkTime: link.lostLinkTime,
            linkQuality: link.linkQuality,
            vtime: link.vtime,
            linkCost: link.linkCost,
            hostname: link.hostname,
            previousLinkStatus: link.previousLinkStatus,
            currentLinkStatus: link.currentLinkStatus,
            neighborLinkQuality: link.neighborLinkQuality,
            symmetryTime: link.symmetryTime,
            seqnoValid: link.seqnoValid,
            pending: link.pending,
            lossHelloInterval: link.lossHelloInterval,
            lossMultiplier: link.lossMultiplier,
            hysteresis: link.hysteresis,
            seqno: link.seqno,
            lossTime: link.lossTime,
            validityTime: link.validityTime,
            olsrInterface: link.olsrInterface,
            lastHelloTime: link.lastHelloTime,
            asymmetryTime: link.asymmetryTime,
            hostname,
        };

        // If ifName starts with "tun", then it is a tunnel
        if (link.ifName.startsWith('tun')) {
            newObj.linkType = 'TUN';
        } else {
            console.log(`Unknown link type: ${link.ifName}`);
            newObj.linkType = 'UNKNOWN';
        }

        const realIP = await dns.lookup(hostname);

        linkInfo[realIP.address] = newObj;
    }

    return linkInfo;
}

function getActiveTunnels() {
    // Check the interfaces for names starting with "tun"
    // Return an integer count of active tunnels

    let count = 0;
    for (const [key, _value] of Object.entries(os.networkInterfaces())) {
        if (key.startsWith('tun')) {
            count += 1;
        }
    }
    return count;
}

getLinkInfo().then((data) => {
    obj.link_info = data;
    console.log(JSON.stringify(obj, null, 2));
}).catch((err) => {
    console.log(err);
});
