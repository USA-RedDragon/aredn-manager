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
    <template #header v-if="$props.admin">
      <div class="table-header-container">
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
    <Column field="active" header="Connected">
      <template #body="slotProps">
        <PVBadge v-if="slotProps.data.active" value="✔️" severity="success"></PVBadge>
        <PVBadge v-else value="✖️" severity="danger"></PVBadge>
        &nbsp;<span v-if="slotProps.data.connection_time == 'Never'">{{slotProps.data.connection_time}}</span>
        <span v-else>{{slotProps.data.connection_time.fromNow()}}</span>
      </template>
    </Column>
    <Column field="hostname" header="Name"></Column>
    <Column field="ip" header="IP"></Column>
    <Column field="password" header="Password" v-if="$props.admin"></Column>
    <Column field="rx_bytes_per_sec" header="Bandwidth Usage" v-if="!$props.admin">
      <template #body="slotProps">
        <p><span style="font-weight: bold;">RX:</span> {{prettyBytes(slotProps.data.rx_bytes_per_sec)}}/s</p>
        <p><span style="font-weight: bold;">TX:</span> {{prettyBytes(slotProps.data.tx_bytes_per_sec)}}/s</p>
      </template>
    </Column>
    <Column field="rx_bytes" header="Session Traffic" v-if="!$props.admin">
      <template #body="slotProps">
        <p><span style="font-weight: bold;">RX:</span> {{prettyBytes(slotProps.data.rx_bytes)}}</p>
        <p><span style="font-weight: bold;">TX:</span> {{prettyBytes(slotProps.data.tx_bytes)}}</p>
      </template>
    </Column>
    <Column field="total_rx_mb" header="Total Traffic" v-if="!$props.admin">
      <template #body="slotProps">
        <p><span style="font-weight: bold;">RX:</span> {{prettyBytes(slotProps.data.total_rx_mb)}}</p>
        <p><span style="font-weight: bold;">TX:</span> {{prettyBytes(slotProps.data.total_tx_mb)}}</p>
      </template>
    </Column>
    <Column field="created_at" header="Created" v-if="$props.admin">
      <template #body="slotProps">{{
        slotProps.data.created_at.fromNow()
      }}</template>
    </Column>
    <template #expansion="slotProps" v-if="$props.admin">
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
import Badge from 'primevue/badge';
import Column from 'primevue/column';
import DataTable from 'primevue/datatable';

import moment from 'moment';
import prettyBytes from 'pretty-bytes';

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
    PVBadge: Badge,
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
    prettyBytes(bytes) {
      if (!bytes) {
        return '0 B';
      }
      return prettyBytes(bytes);
    },
    onPage(event) {
      this.loading = true;
      this.fetchData(event.page + 1, event.rows);
    },
    fetchData(page = 1, limit = 10) {
      this.loading = true;
      API.get(`/tunnels?page=${page}&limit=${limit}&admin=${this.$props.admin}`)
        .then((res) => {
          if (!res.data.tunnels) {
            res.data.tunnels = [];
          }
          for (let i = 0; i < res.data.tunnels.length; i++) {
            res.data.tunnels[i].created_at = moment(
              res.data.tunnels[i].created_at,
            );
            if (res.data.tunnels[i].connection_time == '0001-01-01T00:00:00Z') {
              res.data.tunnels[i].connection_time = 'Never';
            } else {
              res.data.tunnels[i].connection_time = moment(
                res.data.tunnels[i].connection_time,
              );
            }
            if (res.data.tunnels[i].total_rx_mb != 0) {
              // Truncate to 2 decimal places
              res.data.tunnels[i].total_rx_mb = Math.round(res.data.tunnels[i].total_rx_mb * 100) / 100;
              // Convert to bytes
              res.data.tunnels[i].total_rx_mb = res.data.tunnels[i].total_rx_mb * 1024 * 1024;
            }
            if (res.data.tunnels[i].total_tx_mb != 0) {
              // Truncate to 2 decimal places
              res.data.tunnels[i].total_tx_mb = Math.round(res.data.tunnels[i].total_tx_mb * 100) / 100;
              // Convert to bytes
              res.data.tunnels[i].total_tx_mb = res.data.tunnels[i].total_tx_mb * 1024 * 1024;
            }
          }
          this.tunnels = res.data.tunnels;
          this.totalRecords = res.data.total;
          this.loading = false;
        })
        .catch((err) => {
          this.loading = false;
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
