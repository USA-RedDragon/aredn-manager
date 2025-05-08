/// <reference types="vite/client" />
import VueRouter, { Route } from 'vue-router'
import type { ToastMessageOptions } from 'primevue/toast'

/**
 * Toast Service methods.
 *
 * @group Model
 *
 */
export interface ToastServiceMethods {
  /**
   * Displays the message in a suitable Toast component.
   * @param {ToastMessageOptions} message - Message instance.
   */
  add(message: ToastMessageOptions): void
  /**
   * Clears the message.
   * @param {ToastMessageOptions} message - Message instance.
   */
  remove(message: ToastMessageOptions): void
  /**
   * Clears the messages that belongs to the group.
   * @param {string} group - Name of the message group.
   */
  removeGroup(group: string): void
  /**
   * Clears all the messages.
   */
  removeAllGroups(): void
}

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $router: VueRouter
    $route: Route
  }
}
