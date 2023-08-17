export default [
  {
    path: '/',
    name: 'Main',
    component: () => import('../views/MainPage.vue'),
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginPage.vue'),
  },
  {
    path: '/meshes',
    name: 'Meshes',
    component: () => import('../views/MeshesPage.vue'),
  },
  {
    path: '/tunnels',
    name: 'Tunnels',
    component: () => import('../views/TunnelsPage.vue'),
  },
  {
    path: '/nodes',
    name: 'Nodes',
    component: () => import('../views/NodesPage.vue'),
  },
  {
    path: '/admin/users',
    name: 'AdminUsers',
    component: () => import('../views/admin/UsersPage.vue'),
  },
  {
    path: '/admin/users/register',
    name: 'AdminUserRegister',
    component: () => import('../views/admin/RegisterPage.vue'),
  },
  {
    path: '/admin/meshes',
    name: 'AdminMeshes',
    component: () => import('../views/admin/MeshesPage.vue'),
  },
  {
    path: '/admin/tunnels',
    name: 'AdminTunnels',
    component: () => import('../views/admin/TunnelsPage.vue'),
  },
  {
    path: '/admin/tunnels/create',
    name: 'AdminTunnelsCreate',
    component: () => import('../views/admin/TunnelsCreatePage.vue'),
  },
  {
    path: '/admin/meshes/link',
    name: 'AdminMeshLink',
    component: () => import('../views/admin/MeshLinkPage.vue'),
  },
];