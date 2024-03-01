
import { defineInterface } from '@directus/extensions-sdk'
import SelectFunction from './SelectFunction.vue'

export default defineInterface({
  id: 'call-go-select-function',
  name: 'Call Go Select Function',
  icon: 'function',
  description: 'Select a Call Go function.',
  component: SelectFunction,
  options: null,
  types: ['string'],
  system: true,
})
