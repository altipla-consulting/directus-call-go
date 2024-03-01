
import { defineOperationApi } from '@directus/extensions-sdk'
import { listServers } from '../models/servers'

type Options = {
  fnname: string
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
  async handler({ fnname, payload }, { env, logger, data, accountability }) {
    let parts = fnname.split('||')
    if (parts.length !== 2) {
      throw new Error('The function name should be in the format "server||function"')
    }
    if (!listServers().includes(parts[0])) {
      throw new Error(`The server ${parts[0]} is not available`)
    }

    logger = logger.child({ fnname: parts[1], server: parts[0] })

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
      reply = await fetch(`${parts[0]}/__callgo/invoke`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          fnname: parts[1],
          accountability,
          trigger: data.$trigger,
          payload,
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
