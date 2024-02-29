
import { defineOperationApi } from '@directus/extensions-sdk'

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
  id: 'altipla-go-call',
  async handler({ fnname, payload }, { env, logger, data, accountability }) {
    let server = env.ALTIPLA_CALL_GO_URL
    if (!server) {
      logger.error('The ALTIPLA_CALL_GO_URL environment variable is not set')
      throw new Error('The ALTIPLA_CALL_GO_URL environment variable is not set')
    }
    if (server.endsWith('/')) {
      server = server.slice(0, -1)
    }
    
    logger.info({
      msg: 'Call Go function',
      fnname,
      server,
    })

    let reply
    try {
      reply = await fetch(`${server}/__callgo`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
        },
        body: JSON.stringify({
          fnname,
          accountability,
          trigger: data.$trigger,
          payload,
        }),
      })
    } catch (error) {
      logger.error({
        msg: 'The Go function call failed',
        fnname,
        error,
      })
      throw error
    }
    if (!reply.ok) {
      let error = (await reply.text()).trim()

      logger.error({
        msg: 'The Go function call status failed',
        fnname,
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
        fnname,
        error: result.error,
      })
      throw new CallError(result.error)
    }
    return result
  },
})
