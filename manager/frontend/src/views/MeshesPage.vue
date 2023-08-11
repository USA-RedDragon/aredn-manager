<template>
  <div>
    <Card>
      <template #title>Supernode Mesh Connections</template>
      <template #content>
        <MeshesTable />
      </template>
    </Card>
  </div>
</template>

<script>
import Card from 'primevue/card';
import MeshesTable from '../components/MeshesTable.vue';

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
