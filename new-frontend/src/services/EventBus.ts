import mitt from 'mitt'
import type { App } from 'vue'
import type { Emitter, EventType } from 'mitt'

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $EventBus: Emitter<Record<EventType, unknown>>
  }
}

export default {
  install: (app: App<Element>) => {
    app.config.globalProperties.$EventBus = mitt()
  },
}
