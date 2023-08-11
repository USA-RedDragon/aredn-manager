<template>
  <DataTable
    :value="users"
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
      <div class="table-header-container">
        <RouterLink to="/admin/users/register">
          <PVButton
            class="p-button-raised p-button-rounded p-button-success"
            icon="pi pi-plus"
            label="Enroll New Admin"
          />
        </RouterLink>
      </div>
    </template>
    <Column :expander="true" />
    <Column field="id" header="ID"></Column>
    <Column field="username" header="Username"></Column>
    <Column field="created_at" header="Created">
      <template #body="slotProps">{{
        slotProps.data.created_at.fromNow()
      }}</template>
    </Column>
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
import { useUserStore, useSettingsStore } from '@/store';

import API from '@/services/API';

export default {
  name: 'UserTable',
  props: {
  },
  components: {
    PVButton: Button,
    DataTable,
    Column,
  },
  data: function() {
    return {
      users: [],
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
      API.get(`/users?page=${page}&limit=${limit}`)
        .then((res) => {
          if (!res.data.users) {
            res.data.users = [];
          }
          for (let i = 0; i < res.data.users.length; i++) {
            res.data.users[i].created_at = moment(
              res.data.users[i].created_at,
            );
          }
          this.users = res.data.users;
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
        detail: `Users cannot be edited yet.`,
        life: 3000,
      });
    },
    deleteUser(user) {
      if (user.id != 1) {
        this.$confirm.require({
          message: 'Are you sure you want to delete this user?',
          header: 'Delete User',
          icon: 'pi pi-exclamation-triangle',
          acceptClass: 'p-button-danger',
          accept: () => {
            API.delete('/users/' + user.id)
              .then((_res) => {
                this.$toast.add({
                  summary: 'Confirmed',
                  severity: 'success',
                  detail: `User ${user.id} deleted`,
                  life: 3000,
                });
                this.fetchData();
              })
              .catch((err) => {
                console.error(err);
                this.$toast.add({
                  severity: 'error',
                  summary: 'Error',
                  detail: `Error deleting user ${user.id}`,
                  life: 3000,
                });
              });
          },
          reject: () => {},
        });
      } else {
        this.$toast.add({
          summary: 'Cannot delete system account.',
          severity: 'error',
          detail: `The system account cannot be deleted.`,
          life: 3000,
        });
      }
    },
  },
  computed: {
    ...mapStores(useUserStore, useSettingsStore),
  },
};
</script>

<style scoped></style>
