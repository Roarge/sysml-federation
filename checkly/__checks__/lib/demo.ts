// demoUrl names the demo a check session runs against. The checks never
// carry an address of their own: the tunnel or the local port is handed in
// through the environment, once, and read here.

export function demoUrl(): string {
  const url = process.env.DEMO_URL
  if (url === undefined || url === '') {
    throw new Error('DEMO_URL is not set: point it at a running demo, for example http://localhost:8080')
  }
  return url.replace(/\/+$/, '')
}
