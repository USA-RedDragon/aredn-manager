<template>
  <Badge
    :class="status ? 'success':''"
    :variant="status ? '':'destructive'"
  >
    <Check v-if="status" />
    <X v-else />
    {{ status ? '&nbsp;Running':'&nbsp;Stopped' }}
  </Badge>
</template>

<script>
import { Badge } from '@/components/ui/badge';

import {
  Check,
  X,
} from 'lucide-vue-next';


export default {
  components: {
    Badge,
    Check,
    X,
  },
  props: {
    status: {
      type: Boolean,
      default: false,
    },
  },
  data: function() {
    return {};
  },
  mounted() {
  },
  unmounted() {
  },
};
</script>

<style scoped>
.success {
  background-color: var(--success) !important;
}
</style>
