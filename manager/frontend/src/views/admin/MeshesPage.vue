<template>
  <div>
    <PVToast />
    <ConfirmDialog></ConfirmDialog>
    <Card>
      <template #title>Manage Meshes</template>
      <template #content>
        <MeshesTable admin />
      </template>
    </Card>
  </div>
</template>

<script>
import Card from 'primevue/card';
import MeshesTable from '../../components/MeshesTable.vue';

export default {
  components: {
    Card,
    MeshesTable,
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
