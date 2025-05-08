<template>
  <div>
    <span style="display: flex; justify-content: space-evenly">
      <Card style="width: 48%;">
        <CardHeader>
          <CardTitle>Daemon Status</CardTitle>
        </CardHeader>
        <CardContent>
          <table>
            <tbody>
              <tr>
                <td style="width: 7em;">
                  <p style="font-weight: bold;">VTun Daemon</p>
                </td>
                <td>
                  <StatusBadge :status="vtundRunning" />
                </td>
              </tr>
              <tr>
                <td>
                  <p style="font-weight: bold;">Babel Daemon</p>
                </td>
                <td>
                  <StatusBadge :status="babelRunning" />
                </td>
              </tr>
              <tr>
                <td>
                  <p style="font-weight: bold;">OLSR Daemon</p>
                </td>
                <td>
                  <StatusBadge :status="olsrdRunning" />
                </td>
              </tr>
              <tr>
                <td>
                  <p style="font-weight: bold;">DNSMasq</p>
                </td>
                <td>
                  <StatusBadge :status="dnsRunning" />
                </td>
              </tr>
            </tbody>
          </table>
        </CardContent>
      </Card>
      <Card style="width: 48%;">
        <CardHeader>
          <CardTitle>Network Statistics</CardTitle>
        </CardHeader>
        <CardContent>
          <h3 style="font-weight: bold;">Tunnels Connected</h3>
          <p><span style="font-weight: bold;">VTun:</span> {{ vtunTunnelsConnected }}/{{ totalVtunTunnels }}</p>
          <p>
            <span style="font-weight: bold;">Wireguard:</span>
            {{ wireguardTunnelsConnected }}/{{ totalWireguardTunnels }}
          </p>
          <br />
          <h3 style="font-weight: bold;">Current Bandwidth</h3>
          <p><span style="font-weight: bold;">RX:</span> {{ prettyBytes(stats.total_rx_bytes_per_sec) }}/s</p>
          <p><span style="font-weight: bold;">TX:</span> {{ prettyBytes(stats.total_tx_bytes_per_sec) }}/s</p>
          <br />
          <h3 style="font-weight: bold;">Total Traffic Since Restart</h3>
          <p><span style="font-weight: bold;">RX:</span> {{ prettyBytes(stats.total_rx_mb) }}</p>
          <p><span style="font-weight: bold;">TX:</span> {{ prettyBytes(stats.total_tx_mb) }}</p>
        </CardContent>
      </Card>
    </span>
  </div>
</template>

<script>
import { Badge } from '@/components/ui/badge';
import StatusBadge from '@/components/StatusBadge.vue';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';

import prettyBytes from 'pretty-bytes';

import API from '@/services/API';

export default {
  components: {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    StatusBadge,
    Badge,
  },
  created() {
    this.fetchData();
  },
  mounted() {
    this.$EventBus.on('tunnel_disconnection', this.tunnelDisconnected);
    this.$EventBus.on('tunnel_connection', this.tunnelConnected);
    this.$EventBus.on('total_traffic', this.totalTraffic);
    this.$EventBus.on('total_bandwidth', this.totalBandwidth);
  },
  unmounted() {
    this.$EventBus.off('tunnel_disconnection', this.tunnelDisconnected);
    this.$EventBus.off('tunnel_connection', this.tunnelConnected);
    this.$EventBus.off('total_traffic', this.totalTraffic);
    this.$EventBus.off('total_bandwidth', this.totalBandwidth);
  },
  data: function() {
    return {
      babelRunning: true,
      vtundRunning: true,
      olsrdRunning: true,
      dnsRunning: true,
      vtunTunnelsConnected: 0,
      totalVtunTunnels: 0,
      wireguardTunnelsConnected: 0,
      totalWireguardTunnels: 0,
      stats: {},
    };
  },
  methods: {
    tunnelDisconnected(_) {
      this.vtunTunnelsConnected--;
    },
    tunnelConnected(_) {
      this.vtunTunnelsConnected++;
    },
    totalBandwidth(event) {
      if ('TX' in event) {
        this.stats.total_rx_bytes_per_sec = event.RX;
        this.stats.total_tx_bytes_per_sec = event.TX;
      }
    },
    totalTraffic(event) {
      // Truncate to 2 decimal places
      let rx = Math.round(event.RX * 100) / 100;
      // Convert to bytes
      rx = rx * 1024 * 1024;
      this.stats.total_rx_mb = rx;

      // Truncate to 2 decimal places
      let tx = Math.round(event.TX * 100) / 100;
      // Convert to bytes
      tx = tx * 1024 * 1024;
      this.stats.total_tx_mb = tx;
    },
    prettyBytes(bytes) {
      if (!bytes) {
        return '0 B';
      }
      return prettyBytes(bytes);
    },
    fetchData() {
      API.get('/olsr/running').then((res) => {
        this.olsrdRunning = res.data.running;
      });
      API.get('/vtun/running').then((res) => {
        this.vtundRunning = res.data.running;
      });
      API.get('/babel/running').then((res) => {
        this.babelRunning = res.data.running;
      });
      API.get('/dns/running').then((res) => {
        this.dnsRunning = res.data.running;
      });
      API.get('/tunnels/vtun/count/connected').then((res) => {
        this.vtunTunnelsConnected = res.data.count;
      });
      API.get('/tunnels/vtun/count').then((res) => {
        this.totalVtunTunnels = res.data.count;
      });
      API.get('/tunnels/wireguard/count/connected').then((res) => {
        this.wireguardTunnelsConnected = res.data.count;
      });
      API.get('/tunnels/wireguard/count').then((res) => {
        this.totalWireguardTunnels = res.data.count;
      });
      API.get('/stats').then((res) => {
        if (typeof res.data == 'string') {
          return;
        }
        if (!('stats' in res.data)) {
          return;
        }
        if (res.data.stats.total_rx_mb != 0) {
          // Truncate to 2 decimal places
          res.data.stats.total_rx_mb = Math.round(res.data.stats.total_rx_mb * 100) / 100;
          // Convert to bytes
          res.data.stats.total_rx_mb = res.data.stats.total_rx_mb * 1024 * 1024;
        }
        if (res.data.stats.total_tx_mb != 0) {
          // Truncate to 2 decimal places
          res.data.stats.total_tx_mb = Math.round(res.data.stats.total_tx_mb * 100) / 100;
          // Convert to bytes
          res.data.stats.total_tx_mb = res.data.stats.total_tx_mb * 1024 * 1024;
        }
        this.stats = res.data.stats;
      });
    },
  },
  computed: {},
};
</script>

<style scoped></style>
