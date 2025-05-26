interface ToastOptions {
  summary: string
  detail: string
  life: number
  severity: 'success' | 'info' | 'warn' | 'error'
}

interface ToastServiceMethods {
  add(opts: ToastOptions): void
}

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $toast: ToastServiceMethods
  }
}
