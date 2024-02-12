<template>
  <header>
    <h1>
      <RouterLink to="/">AREDN Cloud Node Console</RouterLink>
    </h1>
    <div class="wrapper">
      <nav>
        <RouterLink to="/">Home</RouterLink>
        <RouterLink to="/nodes">Nodes</RouterLink>
        <RouterLink to="/tunnels">Tunnels</RouterLink>

        <router-link
          v-if="this.userStore.loggedIn"
          to="#"
          custom
        >
          <a
            href="#"
            @click="toggleAdminMenu"
            :class="{
              adminNavLink: true,
              'router-link-active': this.$route.path.startsWith('/admin'),
            }"
            >Admin</a
          >
        </router-link>
        <PVMenu
          v-if="this.userStore.loggedIn"
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
        <RouterLink v-if="!this.userStore.loggedIn" to="/login"
          >Login</RouterLink
        >
        <a v-else href="#" @click="logout()">Logout</a>
      </nav>
    </div>
  </header>
</template>

<script>
import Menu from 'primevue/menu';
import API from '@/services/API';

import { mapStores } from 'pinia';
import { useUserStore } from '@/store';

export default {
  components: {
    PVMenu: Menu,
  },
  data: function() {
    return {};
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
    toggleAdminMenu(event) {
      this.$refs.adminMenu.toggle(event);
    },
    toggleTalkgroupsMenu(event) {
      this.$refs.talkgroupsMenu.toggle(event);
    },
  },
  computed: {
    ...mapStores(useUserStore),
  },
};
</script>

<style scoped>
header {
  text-align: center;
  max-height: 100vh;
}

header a,
.adminNavLink {
  text-decoration: none;
}

header a,
header a:visited,
header a:link,
.adminNavLink:visited,
.adminNavLink:link {
  color: var(--primary-text-color);
}

header nav .router-link-active,
.adminNavLink.router-link-active {
  color: var(--cyan-300) !important;
}

header nav a:active,
.adminNavLink:active {
  color: var(--cyan-500) !important;
}

header h1 a,
header h1 a:hover,
header h1 a:visited {
  color: var(--primary-text-color);
  background-color: inherit;
}

nav {
  width: 100%;
  font-size: 1rem;
  text-align: center;
  margin-top: 0.5rem;
  margin-bottom: 1rem;
}

nav a {
  display: inline-block;
  padding: 0 1rem;
  border-left: 2px solid #444;
}

nav a:first-of-type {
  border: 0;
}
</style>
