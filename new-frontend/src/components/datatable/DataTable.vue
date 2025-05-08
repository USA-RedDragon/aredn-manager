<script setup lang="ts" generic="TData, TValue">
import type { ColumnDef, TableOptionsWithReactiveData } from '@tanstack/vue-table'
import {
  FlexRender,
  getCoreRowModel,
  useVueTable,
} from '@tanstack/vue-table'
import { toRefs, onMounted } from 'vue'
import DataTablePagination from './DataTablePagination.vue'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

const props = defineProps<{
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  pagination?: boolean
  rowCount?: number
  pageCount?: number
  fetchData: (_pageIndex: number, _pageSize: number) => Promise<void>
}>()

const { data, rowCount, pageCount } = toRefs(props)

const options: TableOptionsWithReactiveData<TData> = {
  get data() { return data.value },
  get columns() { return props.columns },
  get pageCount() { return pageCount.value },
  get rowCount() { return rowCount.value },
  getCoreRowModel: getCoreRowModel(),
  manualPagination: props.pagination,
  onPaginationChange: (updater) => {
    if (typeof updater === 'function') {
      const newState = updater(table.getState().pagination)
      props.fetchData(newState.pageIndex + 1, newState.pageSize)
    } else {
      props.fetchData(updater.pageIndex + 1, updater.pageSize)
    }
  },
}

const table = useVueTable(options)

onMounted(() => {
  props.fetchData(1, 10)
})
</script>

<template>
  <div>
    <div class="border rounded-md">
      <Table>
        <TableHeader>
          <TableRow v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
            <TableHead v-for="header in headerGroup.headers" :key="header.id">
              <FlexRender
                v-if="!header.isPlaceholder" :render="header.column.columnDef.header"
                :props="header.getContext()"
              />
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <template v-if="table.getRowModel().rows?.length">
            <TableRow
              v-for="row in table.getRowModel().rows" :key="row.id"
              :data-state="row.getIsSelected() ? 'selected' : undefined"
            >
              <TableCell v-for="cell in row.getVisibleCells()" :key="cell.id">
                <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
              </TableCell>
            </TableRow>
          </template>
          <template v-else>
            <TableRow>
              <TableCell :colspan="columns.length" class="h-24 text-center">
                No results.
              </TableCell>
            </TableRow>
          </template>
        </TableBody>
      </Table>
    </div>
    <DataTablePagination v-if="props.pagination" :table="table" />
  </div>
</template>
