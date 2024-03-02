
import { defineOperationApp } from '@directus/extensions-sdk'

export default defineOperationApp({
  id: 'call-go',
  name: 'Go Function Call',
  icon: 'function',
  description: 'Call a Go function running inside an internal app.',
  overview({ fnname, server }) {
    return [
      {
        label: 'Function Name',
        text: fnname,
      },
      {
        label: 'Server',
        text: server,
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
      field: 'server',
      name: 'Server',
      type: 'string',
      meta: {
        hidden: true,
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
