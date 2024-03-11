
import { defineEndpoint } from '@directus/extensions-sdk'
import { asyncHandler } from '@altipla/express-async-handler'
import { findServer, listServers } from '../models/servers'

export default defineEndpoint({
  id: 'call-go',
  handler: (router, { services, getSchema }) => {
    async function load() {
      let { TranslationsService } = services
      let translations = new TranslationsService({
        schema: await getSchema(),
      })
      if (!await translations.translationKeyExists('validationError.callgo', 'es-ES')) {
        await translations.createOne({ key: 'validationError.callgo', language: 'es-ES', value: '{message}' })
        await translations.createOne({ key: 'validationError.callgo', language: 'en-US', value: '{message}' })
        await translations.createOne({ key: 'errors.INVALID', language: 'es-ES', value: 'Error' })
        await translations.createOne({ key: 'errors.INVALID', language: 'en-US', value: 'Error' })
      }
    }
    void load()

    router.get('/servers', (req: any, res) => {
      if (!req.accountability?.admin == null) {
        res.status(403)
        return res.send(`permission denied`)
      }
      
      res.send(listServers())
    })

    router.get('/functions', asyncHandler(async (req: any, res) => {
      if (!req.accountability?.admin) {
        res.status(403).send(`permission denied`)
        return
      }

      if (!req.query.server) {
        res.status(400).send(`server required`)
        return
      }
      let server = findServer(req.query.server)
      if (!server) {
        res.status(400).send(`server ${req.query.server} not found`)
        return
      }

      let headers: Record<string, string> = {}
      if (process.env.ALTIPLA_CALL_GO_TOKEN) {
        headers.Authorization = `Bearer ${process.env.ALTIPLA_CALL_GO_TOKEN}`
      }
      let reply = await fetch(`${server.url}/__callgo/functions`, {
        headers,
      })
      if (!reply.ok) {
        res.status(reply.status).send(await reply.text())
        return
      }
      res.send(await reply.json())
    }))
  },
})
