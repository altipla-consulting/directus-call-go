
import { defineOperationApp } from '@directus/extensions-sdk'

export default defineOperationApp({
  id: 'call-go',
  name: 'Go Function Call',
  icon: 'function',
  description: 'Call a Go function running inside an internal app.',
  overview({ fnname }) {
    return [
      {
        label: 'Function Name',
        text: fnname ? fnname.split('||')[1] : '',
      },
      {
        label: 'Server',
        text: fnname ? fnname.split('||')[0] : '',
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
        interface: 'call-go-select-function',
      },
    },
    {
      field: 'payload',
      name: 'Payload',
      type: 'json',
      meta: {
        width: 'full',
        interface: 'input-code',
        options: {
          language: 'json',
        },
      },
      schema: {
        default_value: '{}',
      },
    },
  ],
})
