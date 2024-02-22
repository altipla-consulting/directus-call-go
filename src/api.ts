
import { defineOperationApi } from '@directus/extensions-sdk'

type Options = {
  fnname: string
}

export default defineOperationApi<Options>({
  id: 'custom',
  handler: ({ fnname }) => {
    console.log({ fnname })
  },
})
