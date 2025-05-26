interface ConfirmOptions {
  message: string
  header: string
  icon: string
  acceptClass: string
  accept: () => void
  reject: () => void
}

interface ConfirmServiceMethods {
  require(opts: ConfirmOptions): void
}

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $confirm: ConfirmServiceMethods
  }
}
