
import { defineOperationApp } from '@directus/extensions-sdk'

export default defineOperationApp({
  id: 'altipla-go-call',
  name: 'Go function call',
  icon: 'hub',
  description: 'Call a Go function running inside an internal app.',
  overview: ({ fnname }) => [
    {
      label: 'Function Name',
      text: fnname,
    },
  ],
  options: [
    {
      field: 'text',
      name: 'Function Name',
      type: 'string',
      meta: {
        width: 'full',
        interface: 'input',
      },
    },
  ],
})
