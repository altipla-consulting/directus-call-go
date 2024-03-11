
import { defineOperationApi } from '@directus/extensions-sdk'
import { findServer } from '../models/servers'
import { createError } from '@directus/errors'

type Options = {
  fnname: string
  server: string
  payload: string
}

let CallError = createError<{ error: string }>('INTERNAL_SERVER_ERROR', ({ error }) => `CallGo failed: ${error}`, 500)

export default defineOperationApi<Options>({
  id: 'call-go',
  async handler({ fnname, server, payload }, { env, logger, data, accountability }) {
    let serverdef = findServer(server)
    if (!serverdef) {
      throw new Error(`The server ${server} is not available`)
    }

    logger = logger.child({ fnname, server: server })

    let headers: Record<string, string> = {
      'Content-Type': 'application/json; charset=utf-8',
    }
    let withAuthorization = false
    if (env.ALTIPLA_CALL_GO_TOKEN) {
      headers.Authorization = `Bearer ${env.ALTIPLA_CALL_GO_TOKEN}`
      withAuthorization = true
    }
    
    logger.info({
      msg: 'CallGo function',
      withAuthorization,
    })

    let reply
    try {
      reply = await fetch(`${serverdef.url}/__callgo/invoke`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          fnname,
          accountability,
          trigger: data.$trigger,
          payload: payload || {},
        }),
      })
    } catch (error) {
      logger.error({
        msg: 'CallGo unvoke failed',
        error,
      })
      throw error
    }
    if (!reply.ok) {
      let error = (await reply.text()).trim()

      logger.error({
        msg: 'CallGo status failed',
        status: reply.status,
        error,
      })
      let terr = new Error(`CallGo failed with status ${reply.status}`)
      terr.cause = error
      throw terr
    }
    let result = await reply.json()
    if (result.error) {
      logger.error({
        msg: 'CallGo internal server error',
        error: result.error,
      })
      throw new CallError(result)
    }
    if (result.callGoError) {
      let custom = createError(result.callGoError.code, result.callGoError.message, result.callGoError.status)
      throw new custom(result.callGoError.extensions)
    }
    return result.payload
  },
})
