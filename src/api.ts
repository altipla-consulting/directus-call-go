
import { defineOperationApi } from '@directus/extensions-sdk'

type Options = {
  fnname: string
}

export default defineOperationApi<Options>({
  id: 'altipla-go-call',
  handler({ fnname }) {
    console.log({ fnname })
  },
})
