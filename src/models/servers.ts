
export function listServers(): string[] {
  let servers = process.env.ALTIPLA_CALL_GO_SERVERS?.split(',')
  if (!servers || !servers.length) {
    throw new Error('The ALTIPLA_CALL_GO_SERVERS environment variable is not set')
  }
  if (servers.some(s => !s.startsWith('http://') && !s.startsWith('https://'))) {
    throw new Error('The ALTIPLA_CALL_GO_SERVERS environment variable should always include the protocol')
  }
  if (servers.some(server => server.endsWith('/'))) {
    throw new Error('The ALTIPLA_CALL_GO_SERVERS environment variable should never end with a slash')
  }

  return servers
}
