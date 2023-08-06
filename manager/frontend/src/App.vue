<template>
  <AppHeader />
  <RouterView />
  <AppFooter />
  <ThemeConfig />
</template>

<script>
import { RouterView } from 'vue-router';
import AppFooter from './components/AppFooter.vue';
import AppHeader from './components/AppHeader.vue';
import ThemeConfig from './components/ThemeConfig.vue';
import API from '@/services/API';

import { mapStores } from 'pinia';
import { useUserStore, useSettingsStore } from '@/store';

export default {
  name: 'App',
  components: {
    RouterView,
    AppHeader,
    AppFooter,
    ThemeConfig,
  },
  data() {
    return {
      socket: null,
    };
  },
  created() {},
  mounted() {
    this.fetchData();
  },
  unmounted() {
  },
  methods: {
    fetchData() {
      // GET /users/me
      API.get('/users/me')
        .then((res) => {
          this.userStore.id = res.data.id;
          this.userStore.username = res.data.username;
          this.userStore.created_at = res.data.created_at;
          this.userStore.loggedIn = true;
        })
        .catch((_err) => {
          this.userStore.loggedIn = false;
        });
    },
  },
  computed: {
    ...mapStores(useUserStore, useSettingsStore),
  },
};
</script>

<style scoped></style>
