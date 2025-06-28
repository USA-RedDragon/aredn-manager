<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { onMounted, ref } from 'vue'
import DataTable from './datatable/DataTable.vue'
import API from '../services/API'

import { h } from 'vue'

interface Service {
  url: string
  protocol: string
  name: string
  shouldLink: boolean
}

interface NodeNoChildren {
  hostname: string
  ip: string
  services: Service[]
}

interface Node {
  hostname: string
  ip: string
  services: Service[]
  children: NodeNoChildren[]
}

const columns: ColumnDef<Node>[] = [
  {
    accessorKey: 'hostname',
    header: () => h('div', {  }, 'Name'),
    cell: ({ row }) => {
      const hostname = row.getValue('hostname')
      return h('a', { target: "_blank", href: `http://${hostname}.local.mesh` }, row.getValue('hostname'))
    },
  },
  {
    accessorKey: 'ip',
    header: () => h('div', {  }, 'IP'),
    cell: ({ row }) => {
      return h('p', { }, row.getValue('ip'))
    },
  },
  {
    accessorKey: 'children',
    header: () => h('div', {  }, 'Devices'),
    cell: ({ row }) => {
      const devices = row.getValue('children') as NodeNoChildren[]
      const ret = []
      if (!devices || devices.length === 0) {
        return h('p', { }, '')
      }
      for (let i = 0; i < devices.length; i++) {
        const device = devices[i]
        ret.push(h('p', { }, device.hostname + ' (' + device.ip + ')'))
      }
      return h('div', { }, ret)
    },
  },
  {
    accessorKey: 'services',
    header: () => h('div', {  }, 'Services'),
    cell: ({ row }) => {
      const services = row.getValue('services') as Service[]
      const ret = []
      if (!services || services.length === 0) {
        return h('p', { }, '')
      }
      for (let i = 0; i < services.length; i++) {
        const service = services[i]
        ret.push(h('a', { target: "_blank", href: service.url }, service.name))
        if (i < services.length - 1) {
          ret.push(h('br', { }))
        }
      }
      return h('div', { }, ret)
    },
  },
]

const data = ref<Node[]>([])
const loading = ref(false)
const devicesCount = ref(0)
const nodesCount = ref(0)
const totalRecords = ref(0)
const limit = ref(10)

const props = defineProps<{
  babel?: boolean
}>()

async function fetchData(page=0, limit=10) {
  loading.value = true;
  const api = props.babel ? '/babel' : '/olsr';
  API.get(`${api}/hosts/count`)
    .then((res) => {
      nodesCount.value = res.data.nodes;
      devicesCount.value = res.data.total;
    })
    .catch((err) => {
      console.error(err);
    });
  API.get(`${api}/hosts?page=${page}&limit=${limit}`)
    .then((res) => {
      if (!res.data.nodes) {
        res.data.nodes = [];
      }

      // Iterate through each node's services and each node's child's services
      // and make them a new URL()
      for (let i = 0; i < res.data.nodes.length; i++) {
        const node = res.data.nodes[i];
        if (node.services != null) {
          for (let j = 0; j < node.services.length; j++) {
            const service = node.services[j];
            service.url = new URL(service.url);
            service.url.hostname = service.url.hostname + '.local.mesh';
            node.services[j] = service;
          }
        }
        if (node.children != null) {
          for (let j = 0; j < node.children.length; j++) {
            const child = node.children[j];
            if (child.services != null) {
              for (let k = 0; k < child.services.length; k++) {
                const service = child.services[k];
                service.url = new URL(service.url);
                service.url.hostname = service.url.hostname + '.local.mesh';
                child.services[k] = service;
              }
            }
            node.children[j] = child;
          }
        }
        res.data.nodes[i] = node;
      }

      data.value = res.data.nodes;
      totalRecords.value = res.data.total;
      loading.value = false;
    })
    .catch((err) => {
      console.error(err);
    });
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="mx-auto">
    <DataTable
      :columns="columns"
      :data="data"
      pagination
      :rowCount="totalRecords"
      :pageCount="Math.ceil(totalRecords / limit)"
      :fetchData="fetchData"
    />
  </div>
</template>
