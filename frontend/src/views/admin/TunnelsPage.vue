<template>
  <div>
    <PVToast />
    <ConfirmDialog></ConfirmDialog>
    <Card>
      <template #title>Tunnels</template>
      <template #content>
        <TunnelsTable admin />
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
