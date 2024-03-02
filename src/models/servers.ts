
export interface Server {
  alias: string
  url: string
}

export function listServers(): Server[] {
  let servers = process.env.ALTIPLA_CALL_GO_SERVERS
    ?.split(',')
    .map(server => {
      if (server.includes('=')) {
        let [alias, url] = server.split('=')
        return { alias, url }
      }
      return { alias: server, url: server }
    })
  if (!servers || !servers.length) {
    throw new Error('The ALTIPLA_CALL_GO_SERVERS environment variable is not set')
  }

  if (servers.some(s => !s.url.startsWith('http://') && !s.url.startsWith('https://'))) {
    throw new Error('The ALTIPLA_CALL_GO_SERVERS environment variable should always include the protocol')
  }
  if (servers.some(server => server.url.endsWith('/'))) {
    throw new Error('The ALTIPLA_CALL_GO_SERVERS environment variable should never end with a slash')
  }

  return servers
}

export function findServer(alias: string): Server | undefined {
  return listServers().find(s => s.alias === alias)
}
