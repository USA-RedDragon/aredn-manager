<template>
  <DataTable
    :value="tunnels"
    v-model:expandedRows="expandedRows"
    v-model:filters="filters"
    dataKey="id"
    :lazy="true"
    :paginator="true"
    :rows="10"
    :totalRecords="totalRecords"
    :loading="loading"
    filterDisplay="menu"
    :globalFilterFields="['hostname']"
    :scrollable="true"
    @page="onPage($event)"
  >
    <template #header>
      <div class="table-header-container">
        <RouterLink v-if="this.admin" :to="'/admin/tunnels/create/' + ($props.wireguard ? 'wireguard':'vtun')">
          <PVButton
            class="p-button-raised p-button-rounded p-button-success"
            icon="pi pi-plus"
            label="New Tunnel"
          />
        </RouterLink>
        <br v-if="this.admin" />
        <div class="flex justify-content-between">
            <PVButton type="button" icon="pi pi-filter-slash" label="Clear" outlined @click="clearFilter()" />
            <span class="p-input-icon-left">
                <i class="pi pi-search" />
                <InputText v-model="filters['global'].value" @change="onFilter()" placeholder="Search" />
            </span>
        </div>
      </div>
    </template>
    <template #empty> No tunnels found. </template>
    <template #loading> Loading tunnels, please wait. </template>
    <Column :expander="true" v-if="$props.admin" />
    <Column field="enabled" header="Enabled">
      <template #body="slotProps">
        <span v-if="slotProps.data.editing">
          <PVCheckbox
              :binary="true"
              v-model="slotProps.data.enabled"
            />
        </span>
        <span v-else>
          <PVBadge v-if="slotProps.data.enabled" value="✔️" severity="success"></PVBadge>
          <PVBadge v-else value="✖️" severity="danger"></PVBadge>
        </span>
      </template>
    </Column>
    <Column field="active" header="Connected">
      <template #body="slotProps">
        <PVBadge v-if="slotProps.data.active" value="✔️" severity="success"></PVBadge>
        <PVBadge v-else value="✖️" severity="danger"></PVBadge>
        &nbsp;<span v-if="slotProps.data.connection_time == 'Never'">{{slotProps.data.connection_time}}</span>
        <span v-else>{{slotProps.data.connection_time.fromNow()}}</span>
      </template>
    </Column>
    <Column field="hostname" header="Name">
      <template #body="slotProps">
        <span v-if="slotProps.data.editing">
          <span class="p-float-label">
            <InputText
              type="text"
              v-model="slotProps.data.hostname"
            />
          </span>
        </span>
        <span v-else>{{slotProps.data.hostname}}</span>
      </template>
    </Column>
    <Column field="ip" header="IP">
      <template #body="slotProps">
        <span v-if="slotProps.data.editing">
          <span class="p-float-label">
            <InputText
              type="text"
              v-model="slotProps.data.ip"
            />
          </span>
        </span>
        <span v-else>{{slotProps.data.ip}}</span>
      </template>
    </Column>
    <Column field="wireguard_port" header="Wireguard Port" v-if="$props.wireguard">
      <template #body="slotProps">
        <span v-if="slotProps.data.wireguard">
          {{slotProps.data.wireguard_port}}
        </span>
        <span v-else>-</span>
      </template>
    </Column>
    <Column field="password" header="Password" v-if="$props.admin">
      <template #body="slotProps">
        <span v-if="slotProps.data.editing">
          <span class="p-float-label">
            <InputText
              type="text"
              v-model="slotProps.data.password"
            />
          </span>
        </span>
        <span v-else>
          <span v-if="slotProps.data.client">
            Private
          </span>
          <span v-else>
            <ClickToCopy :copy="slotProps.data.password" text="Click to copy" />
          </span>
        </span>
      </template>
    </Column>
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
        v-if="!slotProps.data.editing"
        @click="editTunnel(slotProps.data.id)"
      ></PVButton>
      <PVButton
        class="p-button-raised p-button-rounded p-button-primary"
        icon="pi pi-pencil"
        label="Save Changes"
        v-else
        @click="finishEditingTunnel(slotProps.data)"
      ></PVButton>
      <PVButton
        class="p-button-raised p-button-rounded p-button-danger"
        icon="pi pi-trash"
        label="Delete"
        style="margin-left: 0.5em"
        @click="deleteTunnel(slotProps.data)"
      ></PVButton>
    </template>
  </DataTable>
</template>

<script>
import { FilterMatchMode } from 'primevue/api';

import Button from 'primevue/button';
import Badge from 'primevue/badge';
import Column from 'primevue/column';
import Checkbox from 'primevue/checkbox';
import DataTable from 'primevue/datatable';
import InputText from 'primevue/inputtext';

import moment from 'moment';
import prettyBytes from 'pretty-bytes';

import { mapStores } from 'pinia';
import { useSettingsStore } from '@/store';

import ClickToCopy from './ClickToCopy.vue';
import API from '@/services/API';

export default {
  props: {
    admin: {
      type: Boolean,
      default: false,
    },
    wireguard: {
      type: Boolean,
      default: false,
    },
    vtun: {
      type: Boolean,
      default: false,
    },
  },
  components: {
    PVButton: Button,
    PVCheckbox: Checkbox,
    DataTable,
    Column,
    PVBadge: Badge,
    InputText,
    ClickToCopy,
  },
  data: function() {
    return {
      tunnels: [],
      expandedRows: [],
      loading: false,
      totalRecords: 0,
      filters: null,
    };
  },
  created() {
    this.initFilters();
  },
  mounted() {
    this.fetchData();
    this.$EventBus.on('tunnel_stats', this.updateTunnelStats);
    this.$EventBus.on('tunnel_connected', this.updateTunnelConnected);
    this.$EventBus.on('tunnel_disconnected', this.updateTunnelDisconnected);
  },
  unmounted() {
    this.$EventBus.off('tunnel_stats', this.updateTunnelStats);
    this.$EventBus.off('tunnel_connected', this.updateTunnelConnected);
    this.$EventBus.off('tunnel_disconnected', this.updateTunnelDisconnected);
  },
  methods: {
    initFilters() {
      this.filters = {
        global: { value: null, matchMode: FilterMatchMode.CONTAINS },
      };
    },
    onFilter() {
      this.loading = true;
      this.fetchData(this.page, this.filters.global.value);
    },
    clearFilter() {
      this.initFilters();
      this.loading = true;
      this.fetchData(this.page);
    },
    updateTunnelStats(tunnel) {
      for (let i = 0; i < this.tunnels.length; i++) {
        if (this.tunnels[i].id == tunnel.id) {
          this.tunnels[i].rx_bytes_per_sec = tunnel.rx_bytes_per_sec;
          this.tunnels[i].tx_bytes_per_sec = tunnel.tx_bytes_per_sec;
          this.tunnels[i].rx_bytes = tunnel.rx_bytes;
          this.tunnels[i].tx_bytes = tunnel.tx_bytes;
          tunnel.total_rx_mb = tunnel.total_rx_mb * 1024 * 1024;
          this.tunnels[i].total_rx_mb = tunnel.total_rx_mb;
          tunnel.total_tx_mb = tunnel.total_tx_mb * 1024 * 1024;
          this.tunnels[i].total_tx_mb = tunnel.total_tx_mb;
          return;
        }
      }
    },
    updateTunnelConnected(tunnel) {
      for (let i = 0; i < this.tunnels.length; i++) {
        if (this.tunnels[i].id == tunnel.id) {
          this.tunnels[i].active = true;
          if (tunnel.connection_time == '0001-01-01T00:00:00Z' || !tunnel.connection_time) {
            tunnel.connection_time = 'Never';
          } else {
            tunnel.connection_time = moment(tunnel.connection_time);
          }
          this.tunnels[i].connection_time = tunnel.connection_time;
          return;
        }
      }
    },
    updateTunnelDisconnected(tunnel) {
      for (let i = 0; i < this.tunnels.length; i++) {
        if (this.tunnels[i].id == tunnel.id) {
          this.tunnels[i].active = false;
          return;
        }
      }
    },
    prettyBytes(bytes) {
      if (!bytes) {
        return '0 B';
      }
      return prettyBytes(bytes);
    },
    onPage(event) {
      this.loading = true;
      this.fetchData(event.page + 1, this.filters.global.value, event.rows);
    },
    fetchData(page = 1, filter = null, limit = 10) {
      this.loading = true;
      let url = `/tunnels?page=${page}&limit=${limit}&admin=${this.$props.admin}` +
              `&type=${this.$props.wireguard ? 'wireguard' : 'vtun'}`;
      if (filter) {
        url += `&filter=${filter}`;
      }
      API.get(url)
        .then((res) => {
          if (!res.data.tunnels) {
            res.data.tunnels = [];
          }
          for (let i = 0; i < res.data.tunnels.length; i++) {
            res.data.tunnels[i].editing = false;
            res.data.tunnels[i].created_at = moment(
              res.data.tunnels[i].created_at,
            );
            if (res.data.tunnels[i].connection_time == '0001-01-01T00:00:00Z' || !res.data.tunnels[i].connection_time) {
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
    editTunnel(tunnelID) {
      for (let i = 0; i < this.tunnels.length; i++) {
        if (this.tunnels[i].id === tunnelID) {
          this.tunnels[i].editing = true;
          return;
        }
      }
    },
    finishEditingTunnel(tunnel) {
      for (let i = 0; i < this.tunnels.length; i++) {
        if (this.tunnels[i].id === tunnel.id) {
          this.tunnels[i].editing = false;
          break;
        }
      }
      // Send PATCH
      API.patch('/tunnels/', {
        id: tunnel.id,
        hostname: tunnel.hostname,
        password: tunnel.password,
        enabled: tunnel.enabled,
        wireguard: tunnel.wireguard,
        ip: tunnel.ip,
      })
        .then((_res) => {
          this.$toast.add({
            summary: 'Confirmed',
            severity: 'success',
            detail: `Tunnel ${tunnel.id} updated`,
            life: 3000,
          });
        })
        .catch((err) => {
          console.error(err);
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: `Error updating tunnel ${tunnel.id}`,
            life: 3000,
          });
        });
    },
    deleteTunnel(tunnel) {
      this.$confirm.require({
        message: 'Are you sure you want to delete this tunnel?',
        header: 'Delete Tunnel',
        icon: 'pi pi-exclamation-triangle',
        acceptClass: 'p-button-danger',
        accept: () => {
          API.delete('/tunnels/' + tunnel.id)
            .then((_res) => {
              this.$toast.add({
                summary: 'Confirmed',
                severity: 'success',
                detail: `Tunnel ${tunnel.id} deleted`,
                life: 3000,
              });
              this.fetchData();
            })
            .catch((err) => {
              console.error(err);
              this.$toast.add({
                severity: 'error',
                summary: 'Error',
                detail: `Error deleting tunnel ${tunnel.id}`,
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
