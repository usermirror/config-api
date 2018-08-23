const axios = require('axios')
const { getOptions, getPath } = require('./utils')

function ConfigClient(opts) {
  const { host, version, timeout, namespaceId, userAgent } = getOptions(opts)

  this.namespaceId = namespaceId
  this.apiHost = host

  const isBrowser = typeof window !== 'undefined'
  const userAgentHeader = isBrowser ? 'x-user-agent' : 'user-agent'

  const baseURL = host + '/v' + version

  this.client = axios.create({
    baseURL,
    timeout,
    headers: { [userAgentHeader]: userAgent }
  })
}

ConfigClient.prototype.get = function getConfig(opts) {
  if (!opts) {
    return Promise.reject(new Error('config-client.get: missing opts'))
  }

  const namespaceId = opts.namespaceId || this.namespaceId
  const configId = opts.configId

  const response = this.client({
    method: 'get',
    url: getPath({ namespaceId, configId })
  })

  if (!opts.raw) {
    return response.then(r => r.data)
  }

  return response
}

ConfigClient.prototype.set = function setConfig(opts) {
  if (!opts) {
    return Promise.reject(new Error('config-client.set: missing opts'))
  }

  const namespaceId = opts.namespaceId || this.namespaceId
  const configId = opts.configId
  const data = opts.data

  return this.client({
    method: 'put',
    url: getPath({ namespaceId, configId }),
    data
  })
}

ConfigClient.prototype.list = function(opts) {
  // TODO: list configurations
  return Promise.resolve({ type: 'not_found' })
}

module.exports = ConfigClient
