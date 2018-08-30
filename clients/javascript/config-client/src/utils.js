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

module.exports = {
  getOptions,
  getPath
}