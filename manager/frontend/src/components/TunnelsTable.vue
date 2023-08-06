<template>
  <DataTable
    :value="tunnels"
    v-model:expandedRows="expandedRows"
    dataKey="id"
    :lazy="true"
    :paginator="true"
    :rows="10"
    :totalRecords="totalRecords"
    :loading="loading"
    :scrollable="true"
    @page="onPage($event)"
  >
    <template #header>
      <div class="table-header-container" v-if="$props.admin">
        <RouterLink to="/admin/tunnels/create">
          <PVButton
            class="p-button-raised p-button-rounded p-button-success"
            icon="pi pi-plus"
            label="New Tunnel"
          />
        </RouterLink>
      </div>
    </template>
    <Column :expander="true" v-if="$props.admin" />
    <Column field="hostname" header="Name"></Column>
    <Column field="ip" header="IP"></Column>
    <Column field="password" header="Password" v-if="$props.admin"></Column>
    <Column field="created_at" header="Created">
      <template #body="slotProps">{{
        slotProps.data.created_at.fromNow()
      }}</template>
    </Column>
    <Column field="active" header="Active"></Column>
    <template #expansion="slotProps">
      <PVButton
        class="p-button-raised p-button-rounded p-button-primary"
        icon="pi pi-pencil"
        label="Edit"
        @click="editUser(slotProps.data)"
      ></PVButton>
      <PVButton
        class="p-button-raised p-button-rounded p-button-danger"
        icon="pi pi-trash"
        label="Delete"
        style="margin-left: 0.5em"
        @click="deleteUser(slotProps.data)"
      ></PVButton>
    </template>
  </DataTable>
</template>

<script>
import Button from 'primevue/button';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';

import moment from 'moment';

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
    PVButton: Button,
    DataTable,
    Column,
  },
  data: function() {
    return {
      tunnels: [],
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
      this.loading = true;
      this.fetchData(event.page + 1, event.rows);
    },
    fetchData(page = 1, limit = 10) {
      API.get(`/tunnels?page=${page}&limit=${limit}&admin=${this.$props.admin}`)
        .then((res) => {
          if (!res.data.tunnels) {
            res.data.tunnels = [];
          }
          for (let i = 0; i < res.data.tunnels.length; i++) {
            res.data.tunnels[i].created_at = moment(
              res.data.tunnels[i].created_at,
            );
          }
          this.tunnels = res.data.tunnels;
          this.totalRecords = res.data.total;
          this.loading = false;
        })
        .catch((err) => {
          console.error(err);
        });
    },
    editUser(_user) {
      this.$toast.add({
        summary: 'Not Implemented',
        severity: 'error',
        detail: `Tunnels cannot be edited yet.`,
        life: 3000,
      });
    },
    deleteUser(user) {
      this.$confirm.require({
        message: 'Are you sure you want to delete this tunnel?',
        header: 'Delete Tunnel',
        icon: 'pi pi-exclamation-triangle',
        acceptClass: 'p-button-danger',
        accept: () => {
          API.delete('/tunnels/' + user.id)
            .then((_res) => {
              this.$toast.add({
                summary: 'Confirmed',
                severity: 'success',
                detail: `Tunnel ${user.id} deleted`,
                life: 3000,
              });
              this.fetchData();
            })
            .catch((err) => {
              console.error(err);
              this.$toast.add({
                severity: 'error',
                summary: 'Error',
                detail: `Error deleting tunnel ${user.id}`,
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
