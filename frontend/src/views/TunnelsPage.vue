<template>
  <div>
    <Card>
      <template #title>VTun Tunnels</template>
      <template #content>
        <TunnelsTable vtun />
      </template>
    </Card>
    <br />
    <Card>
      <template #title>Wireguard Tunnels</template>
      <template #content>
        <TunnelsTable wireguard />
      </template>
    </Card>
  </div>
</template>

<script>
import Card from 'primevue/card';
import TunnelsTable from '@/components/TunnelsTable.vue';

export default {
  components: {
    Card,
    TunnelsTable,
  },
  created() {},
  mounted() {},
  unmounted() {},
  data: function() {
    return {};
  },
  methods: {},
  computed: {},
};
</script>

<style scoped></style>
