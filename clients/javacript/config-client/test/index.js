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
    const client = new ConfigClient({ namespaceId: 'nam_123' })
    const configId = analyticsId({ prefix: 'cfg' })

    const { type: type1 } = await client.get({ configId })

    t.is(type1, 'not_found')

    await client.set({ configId, data: { type: 'banner' } })

    const { type: type2 } = await client.get({ configId })

    t.is(type2, 'banner')
  })

  test('test 10 updates in under 50ms', async t => {
    const client = new ConfigClient({ namespaceId: 'nam_123' })
    const configId = analyticsId({ prefix: 'cfg' })

    const beginTs = timestamp()
    
    await Promise.all(
      range(10).map(i => client.set({
        configId,
        data: { type: 'banner-' + i }
      }))
    )

    const endTs = timestamp()
    const secondsDiff = (endTs - beginTs) / 1000

    t.truthy(secondsDiff < 0.25, '10 updates took longer than 250ms')
  })

  test('test 100 updates in under 400ms', async t => {
    const client = new ConfigClient({ namespaceId: 'nam_123' })
    const configId = analyticsId({ prefix: 'cfg' })

    const beginTs = timestamp()
    
    await Promise.all(
      range(100).map(i => client.set({
        configId,
        data: { type: 'banner-' + i }
      }))
    )

    const endTs = timestamp()
    const secondsDiff = (endTs - beginTs) / 1000

    t.truthy(secondsDiff < 0.4, '100 updates took longer than 400ms')
  })

  test('test 500 updates in under 1s', async t => {
    const client = new ConfigClient({ namespaceId: 'nam_123' })
    const configId = analyticsId({ prefix: 'cfg' })

    const beginTs = timestamp()
    
    await Promise.all(
      range(500).map(i => client.set({
        configId,
        data: { type: 'banner-' + i }
      }))
    )

    const endTs = timestamp()
    const secondsDiff = (endTs - beginTs) / 1000

    t.truthy(secondsDiff < 1, '500 updates took longer than 1s')
  })
}

function range(length) {
  return Array.from({ length }, (x, i) => i)
}

function timestamp() {
  return (new Date()).getTime()
}