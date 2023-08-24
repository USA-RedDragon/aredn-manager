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

import { getWebsocketURI } from '@/services/util';
import ws from '@/services/ws';

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
    this.socket = ws.connect(getWebsocketURI() + '/events', this.onWebsocketMessage);
  },
  unmounted() {
    if (this.socket) {
      this.socket.close();
    }
  },
  methods: {
    onWebsocketMessage(event) {
      const data = JSON.parse(event.data);
      this.$EventBus.emit(data.type, data.data);
    },
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
