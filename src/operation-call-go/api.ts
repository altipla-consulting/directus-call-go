
import { defineOperationApi } from '@directus/extensions-sdk'
import { findServer } from '../models/servers'

type Options = {
  fnname: string
  server: string
  payload: string
}

class CallError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'CallError'
  }
}

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
      msg: 'Call Go function',
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
        msg: 'The Go function call failed',
        error,
      })
      throw error
    }
    if (!reply.ok) {
      let error = (await reply.text()).trim()

      logger.error({
        msg: 'The Go function call status failed',
        status: reply.status,
        error,
      })
      let terr = new Error(`The Go function call failed with status ${reply.status}`)
      terr.cause = error
      throw terr
    }
    let result = await reply.json()
    if (result.error) {
      logger.error({
        msg: 'The Go function call returned an error',
        error: result.error,
      })
      throw new CallError(result.error)
    }
    return result
  },
})
