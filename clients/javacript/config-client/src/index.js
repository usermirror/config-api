const axios = require('axios')

function ConfigClient(opts) {
  const { host, version, timeout, namespaceId, userAgent } = getOptions(opts)

  this.namespaceId = namespaceId
  this.apiHost = host

  const baseURL = host + '/v' + version

  this.client = axios.create({
    baseURL,
    timeout,
    headers: {
      'user-agent': userAgent
    }
  })
}

ConfigClient.prototype.get = function (opts) {
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

ConfigClient.prototype.set = function (opts) {
  if (!opts) {
    return Promise.reject(new Error('config-client.get: missing opts'))
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

ConfigClient.prototype.list = function (opts) {
  // TODO: list configurations
}

const defaultOptions = {
  host: 'http://localhost:8888',
  userAgent: 'config-client@0.1.0',
  timeout: 5000,
  version: 1
}

function getOptions(opts) {
  return Object.assign({}, defaultOptions, opts || {})
}

function getPath(opts) {
  const { namespaceId, configId } = opts

  return ['/namespaces/', namespaceId, '/configs/', configId].join('')
}

module.exports = ConfigClient