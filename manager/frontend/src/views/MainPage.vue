<template>
  <div>
    <Card>
      <template #title>Cloud Node Status</template>
      <template #content>
        <h3>VTun Daemon</h3>
        <p>{{ vtundRunning ? 'Running':'Stopped' }}</p>
        <h3>OLSR Daemon</h3>
        <p>{{ olsrdRunning ? 'Running':'Stopped' }}</p>
        <h3>Tunnels Connected</h3>
        <p>{{ tunnelsConnected }}/{{ totalTunnels }}</p>
      </template>
    </Card>
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
      tunnelsConnected: 0,
      totalTunnels: 0,
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
      API.get('/tunnels').then((res) => {
        this.tunnelsConnected = 0;
        for (const tunnel of res.data.tunnels) {
          if (tunnel.active) {
            this.tunnelsConnected++;
          }
        }
        this.totalTunnels = res.data.total;
      });
    },
  },
  computed: {},
};
</script>

<style scoped></style>
