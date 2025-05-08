<template>
  <div>
    <PVToast />
    <ConfirmDialog></ConfirmDialog>
    <Card>
      <template #title>Manage Admins</template>
      <template #content>
        <UserTable />
      </template>
    </Card>
  </div>
</template>

<script lang="ts">
import Card from 'primevue/card';
import UserTable from '@/components/UserTable.vue';

export default {
  components: {
    Card,
    UserTable,
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
