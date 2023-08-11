<template>
  <DataTable
    :value="meshes"
    dataKey="id"
    :paginator="false"
    v-model:expandedRows="expandedRows"
    :totalRecords="totalRecords"
    :loading="loading"
    :scrollable="true"
    @page="onPage($event)"
  >
    <template #header v-if="$props.admin">
      <div class="table-header-container">
        <RouterLink to="/admin/meshes/link">
          <PVButton
            class="p-button-raised p-button-rounded p-button-success"
            icon="pi pi-plus"
            label="New Mesh Link"
          />
        </RouterLink>
      </div>
    </template>
    <Column :expander="true" v-if="$props.admin" />
    <Column field="name" header="Name"></Column>
    <Column field="ips" header="IP Addresses"></Column>
    <template #expansion="slotProps" v-if="$props.admin">
      <PVButton
        class="p-button-raised p-button-rounded p-button-primary"
        icon="pi pi-pencil"
        label="Edit"
        @click="editMesh(slotProps.data)"
      ></PVButton>
      <PVButton
        class="p-button-raised p-button-rounded p-button-danger"
        icon="pi pi-trash"
        label="Delete"
        style="margin-left: 0.5em"
        @click="deleteMesh(slotProps.data)"
      ></PVButton>
    </template>
  </DataTable>
</template>

<script>
import Button from 'primevue/button';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import { mapStores } from 'pinia';
import { useSettingsStore } from '@/store';

import API from '@/services/API';

export default {
  props: {
    admin: {
      type: Boolean,
      default: false,
    },
  },
  components: {
    DataTable,
    Column,
    PVButton: Button,
  },
  data: function() {
    return {
      meshes: [],
      expandedRows: [],
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
    onPage(event) {
      this.fetchData(event.page + 1, event.rows);
    },
    fetchData(page = 1, limit = 10) {
      this.loading = true;
      API.get(`/meshes?page=${page}&limit=${limit}&admin=${this.$props.admin}`)
        .then((res) => {
          if (!res.data.meshes) {
            res.data.meshes = [];
          }
          this.meshes = res.data.meshes;
          this.totalRecords = res.data.total;
          this.loading = false;
        })
        .catch((err) => {
          this.loading = false;
          console.error(err);
        });
    },
    editMesh(_user) {
      this.$toast.add({
        summary: 'Not Implemented',
        severity: 'error',
        detail: `Mesh links cannot be edited yet.`,
        life: 3000,
      });
    },
    deleteMesh(mesh) {
      this.$confirm.require({
        message: 'Are you sure you want to delete this mesh link?',
        header: 'Delete Mesh Link',
        icon: 'pi pi-exclamation-triangle',
        acceptClass: 'p-button-danger',
        accept: () => {
          API.delete('/mesh/' + mesh.id)
            .then((_res) => {
              this.$toast.add({
                summary: 'Confirmed',
                severity: 'success',
                detail: `Mesh ${mesh.id} deleted`,
                life: 3000,
              });
              this.fetchData();
            })
            .catch((err) => {
              console.error(err);
              this.$toast.add({
                severity: 'error',
                summary: 'Error',
                detail: `Error deleting mesh ${mesh.id}`,
                life: 3000,
              });
            });
        },
        reject: () => {},
      });
    },
  },
  computed: {
    ...mapStores(useSettingsStore),
  },
};
</script>

<style scoped></style>
