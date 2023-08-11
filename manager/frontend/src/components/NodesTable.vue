<template>
  <DataTable
    :value="hosts"
    dataKey="ip"
    :paginator="false"
    :totalRecords="totalRecords"
    :loading="loading"
    :scrollable="true"
    @page="onPage($event)"
  >
    <Column field="hostname" header="Name">
      <template #body="slotProps">
        <a target="_blank" :href="'http://' + slotProps.data.hostname + '.local.mesh'">
          {{ slotProps.data.hostname }}
        </a>
      </template>
    </Column>
    <Column field="ip" header="IP"></Column>
    <Column field="children" header="Devices">
      <template #body="slotProps">
        <p v-for="child in slotProps.data.children" v-bind:key="child.hostname">
          {{ child.hostname }} ({{ child.ip }})
        </p>
      </template>
    </Column>
    <Column field="services" header="Services">
      <template #body="slotProps">
        <span v-for="child in slotProps.data.children" v-bind:key="child.hostname">
          <p v-for="service in child.services" v-bind:key="service.url">
            {{ service.name }}
          </p>
        </span>
        <p v-for="service in slotProps.data.services" v-bind:key="service.url">
          <span v-if="service.should_link">
            <a target="_blank" :href="service.url">{{ service.name }}</a>
          </span>
          <span v-else>{{ service.name }}</span>
        </p>
      </template>
    </Column>
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
      hosts: [],
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
      API.get(`/olsr/hosts`)
        .then((res) => {
          if (!res.data.nodes) {
            res.data.nodes = [];
          }
          this.hosts = res.data.nodes;
          this.totalRecords = res.data.nodes.length;
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
