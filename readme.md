<p align="center">
  <strong>config-api</strong>
</p>

<p align="center">
  <a href="https://travis-ci.org/usermirror/config-api">
    <img src="https://travis-ci.org/usermirror/config-api.svg?branch=master">
  </a>
</p>

<p align="center">
  Simple configuration API service backed by<br/>
  high-performance KV stores and durable SQL databases.
</p>

<br/>

## Installation

You can either run the `config-api` with Go directly:

```
$ go get -u github.com/usermirror/config-api
# start a supported storage backend, e.g. redis
$ config-api --storage-backend redis
```

Or with Docker:

```
$ docker run -it -p 8888:8888 usermirror/config-api
```

## Storage Backends

| Name           | Backend Type       | Supported? |
| -------------- | ------------------ | ---------- |
| Etcd           | Key-value Store    | **Yes!**   |
| Redis          | Key-value Store    | **Yes!**   |
| Vault          | Key-value Store    | **Yes!**   |
| Postgres       | SQL Database       | **Yes!**   |
| Cassandra      | SQL Columnn Store  | Not yet    |
| CockroachDB    | SQL Database       | Not yet    |
| Memcached      | Key-value Store    | Not yet    |
| MySQL          | SQL Database       | Not yet    |
