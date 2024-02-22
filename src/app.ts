
import { defineOperationApp } from '@directus/extensions-sdk'

export default defineOperationApp({
  id: 'altipla-go-call',
  name: 'Go Function Call',
  icon: 'hub',
  description: 'Call a Go function running inside an internal app.',
  overview({ fnname }) {
    return [
      {
        label: 'Function Name',
        text: fnname,
      },
    ]
  },
  options: [
    {
      field: 'fnname',
      name: 'Function Name',
      type: 'string',
      meta: {
        width: 'full',
        interface: 'input',
      },
    },
  ],
})
