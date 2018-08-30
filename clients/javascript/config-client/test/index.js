const test = require('ava')
const analyticsId = require('analytics-id')
const ConfigClient = require('../src')

test('new client without error', t => {
  new ConfigClient()

  t.pass()
})

const shouldTestRealAPI = process.env.REAL || process.env.real

if (shouldTestRealAPI) {
  test('test create', async t => {
    const client = new ConfigClient({ namespaceId: 'nsc_123' })
    const configId = analyticsId({ prefix: 'cfg' })

    const { __type: type1 } = await client.get({ configId })

    t.is(type1, 'not_found')

    await client.set({ configId, data: { type: 'banner' } })

    const { __type: type2 } = await client.get({ configId })

    t.is(type2, 'banner')
  })

  test('test raw create', async t => {
    const client = new ConfigClient({ namespaceId: 'nsc_123' })
    const configId = analyticsId({ prefix: 'cfg' })

    const { type: type1 } = await client.get({ configId, format: 'raw' })

    t.is(type1, 'not_found')

    await client.set({ configId, data: { type: 'banner' } })

    const { type: type2 } = await client.get({ configId, format: 'raw' })

    t.is(type2, 'banner')
  })

  test('test 10 updates in under 50ms', async t => {
    const client = new ConfigClient({ namespaceId: 'nsc_123' })
    const configId = analyticsId({ prefix: 'cfg' })

    const beginTs = timestamp()

    await Promise.all(
      range(10).map(i =>
        client.set({
          configId,
          data: { type: 'banner-' + i }
        })
      )
    )

    const endTs = timestamp()
    const secondsDiff = (endTs - beginTs) / 1000

    t.truthy(secondsDiff < 0.25, '10 updates took longer than 50ms')
  })
}

function range(length) {
  return Array.from({ length }, (x, i) => i)
}

function timestamp() {
  return new Date().getTime()
}
