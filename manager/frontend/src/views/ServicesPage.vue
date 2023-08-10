<template>
  <div>
    <Card>
      <template #title>Services</template>
      <template #content>
        <ServicesTable />
      </template>
    </Card>
  </div>
</template>

<script>
import Card from 'primevue/card';
import ServicesTable from '../components/ServicesTable.vue';

export default {
  components: {
    Card,
    ServicesTable,
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
