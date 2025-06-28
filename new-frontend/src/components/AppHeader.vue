<template>
  <header>
    <h1>
      <RouterLink to="/">Cloud Node Console</RouterLink>
    </h1>
    <nav>
      <RouterLink to="/">Home</RouterLink>
      <RouterLink to="/olsr">OLSR</RouterLink>
      <RouterLink to="/babel">Babel</RouterLink>
      <RouterLink to="/tunnels">Tunnels</RouterLink>
      <RouterLink v-if="hasMeshmap" to="/meshmap">Mesh Map</RouterLink>

      <router-link
        v-if="userStore.loggedIn"
        to="#"
        custom
      >
        <a
          href="#"
          @click="toggleAdminMenu"
          :class="{
            adminNavLink: true,
            'router-link-active': $route.path.startsWith('/admin'),
          }"
          >Admin</a
        >
      </router-link>
      <PVMenu
        v-if="userStore.loggedIn"
        ref="adminMenu"
        :popup="true"
        :model="[
          {
            label: '&nbsp;&nbsp;Tunnels',
            to: '/admin/tunnels',
          },
          {
            label: '&nbsp;&nbsp;Admin Users',
            to: '/admin/users',
          },
        ]"
      >
        <template #item="{ item }">
          <router-link
            :to="item.to"
            custom
            v-slot="{ href, navigate, isActive, isExactActive }"
          >
            <a
              :href="href"
              @click="navigate"
              :class="{
                adminNavLink: true,
                'router-link-active': isActive,
                'router-link-active-exact': isExactActive,
              }"
            >
              <div>{{ item.label }}</div>
            </a>
          </router-link>
        </template>
      </PVMenu>
      <RouterLink v-if="!userStore.loggedIn" to="/login"
        >Login</RouterLink
      >
      <a v-else href="#" @click="logout()">Logout</a>
    </nav>
    <ColorModeButton class="button" />
  </header>
</template>

<script lang="ts">
import Menu from 'primevue/menu';
import API from '@/services/API';
import ColorModeButton from '@/components/ColorModeButton.vue';

import { mapStores } from 'pinia';
import { ref } from 'vue';
import { useUserStore } from '@/store';

const adminMenu = ref<Menu>();

export default {
  components: {
    PVMenu: Menu,
    ColorModeButton,
  },
  data: function() {
    return {
      hasMeshmap: true,
    };
  },
  mounted() {},
  methods: {
    logout() {
      API.get('/auth/logout')
        .then((_res) => {
          this.userStore.loggedIn = false;
          this.$router.push('/login');
        })
        .catch((err) => {
          console.error(err);
        });
    },
    toggleAdminMenu(event: Event) {
      adminMenu?.value?.toggle(event);
    },
  },
  computed: {
    ...mapStores(useUserStore),
  },
};
</script>

<style scoped>
header {
  height: 3em;
  padding: 0.5em;
  margin: auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  background-color: var(--secondary);
}

header h1,
header nav,
.button {
  font-size: 1rem;
  width: 33%;
}

.button {
  text-align: right;
}

header h1,
header nav {
  display: inline;
}

header nav .router-link-active,
.adminNavLink.router-link-active {
  color: var(--secondary-foreground) !important;
  font-weight: bolder;
}

nav {
  text-align: center;
}

nav a {
  padding: 0 1rem;
  border-left: 1px solid #444;
}

nav a:first-of-type {
  border: 0;
}
</style>
