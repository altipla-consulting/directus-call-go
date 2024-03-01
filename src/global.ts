
import { getCurrentInstance } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'

// We cannot access the root app instance in Directus, so this is a hack to install the plugin the first time we use it.
export function installVueQuery() {
  let g = window as any
  if (g.$$pluginVueQuery) {
    return
  }

  getCurrentInstance()?.appContext.app.use(VueQueryPlugin)
  g.$$pluginVueQuery = true
}
