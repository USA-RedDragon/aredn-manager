<template>
  <DataTable
    :value="services"
    dataKey="url"
    :paginator="false"
    :totalRecords="totalRecords"
    :loading="loading"
    :scrollable="true"
    @page="onPage($event)"
  >
    <Column field="name" header="Name">
      <template #body="slotProps">
        <a target="_blank"
          v-if="slotProps.data.should_link"
          :href="slotProps.data.url.protocol + '//' +
            slotProps.data.url.hostname + '.local.mesh' +
            ((slotProps.data.url.port == 0) ? '':(':' + slotProps.data.url.port)) +
            slotProps.data.url.pathname + slotProps.data.url.hash">
          {{ slotProps.data.name }}
        </a>
        <p v-else>{{ slotProps.data.name }}</p>
      </template>

    </Column>
    <Column field="protocol" header="Protocol"></Column>
  </DataTable>
</template>

<script>
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';

import { mapStores } from 'pinia';
import { useSettingsStore } from '@/store';

import API from '@/services/API';

export default {
  props: {},
  components: {
    DataTable,
    Column,
  },
  data: function() {
    return {
      services: [],
      loading: false,
      totalRecords: 0,
    };
  },
  mounted() {
    this.fetchData();
  },
  unmounted() {
  },
  methods: {
    fetchData() {
      this.loading = true;
      API.get(`/olsr/services`)
        .then((res) => {
          if (!res.data.services) {
            res.data.services = [];
          }
          for (let i = 0; i < res.data.services.length; i++) {
            res.data.services[i].url = new URL(res.data.services[i].url);
          }
          this.services = res.data.services;
          this.totalRecords = res.data.services.length;
          this.loading = false;
        })
        .catch((err) => {
          console.error(err);
        });
    },
  },
  computed: {
    ...mapStores(useSettingsStore),
  },
};
</script>

<style scoped></style>
