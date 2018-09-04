const axios = require('axios')
const { getOptions, getPath } = require('./utils')

function ConfigClient(opts) {
  const {
    host,
    timeout,
    version,
    userAgent,
    namespaceId,
    namespaceToken
  } = getOptions(opts)

  this.namespaceId = namespaceId
  this.apiHost = host

  const isBrowser = typeof window !== 'undefined'
  const userAgentHeader = isBrowser ? 'x-user-agent' : 'user-agent'

  const baseURL = host + '/v' + version
  const headers = {}

  if (userAgent) {
    headers[userAgentHeader] = userAgent
  }

  if (namespaceToken) {
    headers.authorization = 'Bearer ' + namespaceToken
  }

  this.client = axios.create({
    baseURL,
    timeout,
    headers
  })
}

ConfigClient.prototype.get = function getConfig(opts) {
  if (!opts) {
    return Promise.reject(new Error('config-client.get: missing opts'))
  }

  const namespaceId = opts.namespaceId || this.namespaceId
  const namespaceToken = opts.namespaceToken || opts.token
  const configId = opts.configId
  const format = opts.format
  const path = opts.path

  const headers = {}

  if (namespaceToken) {
    headers.authorization = 'Bearer ' + namespaceToken
  }

  const response = this.client({
    method: 'get',
    headers,
    url: getPath({ namespaceId, configId, path })
  })

  // return as native Response object
  if (format === 'response') {
    return response
  }

  const dataPromise = response.then(r => r.data)

  // return as raw JSON response from server
  if (format === 'raw') {
    return dataPromise
  }

  // otherwise, format as a config value, { ...body, __type }
  return dataPromise.then(result => {
    const { type, body } = result

    return Object.assign({}, body, { __type: type })
  })
}

ConfigClient.prototype.set = function setConfig(opts) {
  if (!opts) {
    return Promise.reject(new Error('config-client.set: missing opts'))
  }

  const namespaceId = opts.namespaceId || this.namespaceId
  const namespaceToken = opts.namespaceToken || opts.token
  const configId = opts.configId
  const data = opts.data

  const headers = {}

  if (namespaceToken) {
    headers.authorization = 'Bearer ' + namespaceToken
  }

  return this.client({
    method: 'put',
    url: getPath({ namespaceId, configId }),
    headers,
    data
  })
}

ConfigClient.prototype.list = function(opts) {
  if (!opts) {
    if (!this.namespaceId) {
      return Promise.reject(new Error('config-client.list: missing opts'))
    } else {
      opts = {}
    }
  }

  const namespaceId = opts.namespaceId || this.namespaceId
  const namespaceToken = opts.namespaceToken || opts.token
  const format = opts.format
  const path = opts.path || '/configs'

  const headers = {}

  if (namespaceToken) {
    headers.authorization = 'Bearer ' + namespaceToken
  }

  const response = this.client({
    method: 'get',
    url: getPath({ namespaceId, path }),
    headers
  })

  // return as native Response object
  if (format === 'response') {
    return response
  }

  const dataPromise = response.then(r => r.data)

  // return as raw JSON response from server
  if (format === 'raw') {
    return dataPromise
  }

  return dataPromise.then(result => {
    const { type, items } = result

    return {
      type,
      items
    }
  })
}

module.exports = ConfigClient
