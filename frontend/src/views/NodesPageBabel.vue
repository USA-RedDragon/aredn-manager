<template>
  <div>
    <Card>
      <template #title>Babel Nodes</template>
      <template #content>
        <NodesTable babel />
      </template>
    </Card>
  </div>
</template>

<script>
import Card from 'primevue/card';
import NodesTable from '../components/NodesTable.vue';

export default {
  components: {
    Card,
    NodesTable,
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
