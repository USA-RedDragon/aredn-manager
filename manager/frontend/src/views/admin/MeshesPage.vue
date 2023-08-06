<template>
  <div>
    <PVToast />
    <ConfirmDialog></ConfirmDialog>
    <Card>
      <template #title>Manage Meshes</template>
      <template #content>
        <p>Main</p>
      </template>
    </Card>
  </div>
</template>

<script>
import Card from 'primevue/card';

export default {
  components: {
    Card,
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
