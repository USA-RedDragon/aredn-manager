<template>
  <div>
    <span style="display: flex; justify-content: space-evenly">
      <Card style="width: 48%;">
        <template #title>Daemon Status</template>
        <template #content>
          <h3 style="font-weight: bold;">VTun Daemon</h3>
          <p>{{ vtundRunning ? 'Running':'Stopped' }}</p>
          <br />
          <h3 style="font-weight: bold;">OLSR Daemon</h3>
          <p>{{ olsrdRunning ? 'Running':'Stopped' }}</p>
          <br />
          <h3 style="font-weight: bold;">DNSMasq</h3>
          <p>{{ dnsRunning ? 'Running':'Stopped' }}</p>
        </template>
      </Card>
      <Card style="width: 48%;">
        <template #title>Network Statistics</template>
        <template #content>
          <h3 style="font-weight: bold;">Tunnels Connected</h3>
          <p>{{ tunnelsConnected }}/{{ totalTunnels }}</p>
          <br />
          <h3 style="font-weight: bold;">Bytes RX/TX per Second</h3>
          <p>{{ stats.total_rx_bytes_per_sec }} bytes/s</p>
          <p>{{ stats.total_tx_bytes_per_sec }} bytes/s</p>
          <br />
          <h3 style="font-weight: bold;">Total Bytes RX/TX</h3>
          <p>{{ stats.total_rx_mb }} MBytes</p>
          <p>{{ stats.total_tx_mb }} MBytes</p>
        </template>
      </Card>
    </span>
  </div>
</template>

<script>
import Card from 'primevue/card';

import API from '@/services/API';

export default {
  components: {
    Card,
  },
  created() {
    this.fetchData();
  },
  mounted() {},
  unmounted() {},
  data: function() {
    return {
      vtundRunning: true,
      olsrdRunning: true,
      dnsRunning: true,
      tunnelsConnected: 0,
      totalTunnels: 0,
      stats: {},
    };
  },
  methods: {
    fetchData() {
      API.get('/olsr/running').then((res) => {
        this.olsrdRunning = res.data.running;
      });
      API.get('/vtun/running').then((res) => {
        this.vtundRunning = res.data.running;
      });
      API.get('/dns/running').then((res) => {
        this.dnsRunning = res.data.running;
      });
      API.get('/tunnels').then((res) => {
        this.tunnelsConnected = 0;
        if ('tunnels' in res.data == false || res.data.tunnels == undefined || res.data.tunnels == null) {
          return;
        }
        for (const tunnel of res.data.tunnels) {
          if (tunnel.active) {
            this.tunnelsConnected++;
          }
        }
        this.totalTunnels = res.data.total;
      });
      API.get('/stats').then((res) => {
        if (res.data.stats.total_rx_mb != 0) {
          // Truncate to 2 decimal places
          res.data.stats.total_rx_mb = Math.round(res.data.stats.total_rx_mb * 100) / 100;
        }
        if (res.data.stats.total_tx_mb != 0) {
          // Truncate to 2 decimal places
          res.data.stats.total_tx_mb = Math.round(res.data.stats.total_tx_mb * 100) / 100;
        }
        this.stats = res.data.stats;
      });
    },
  },
  computed: {},
};
</script>

<style scoped></style>
