# graphiteauthd: Graphite Authentication Proxy

Drops any metrics where the root namespace does not match a sha256 API key.

Simple proof of concept TCP proxy for Graphite metrics, use in production at
your own risk ;).

# Usage

```
Usage of graphiteauthd:
  -apikey="MYAPIKEY": API key to accept
  -beconnections=1: Number of concurrent connections to open to the Graphite backend
  -colour=true: Colourise output
  -listen=":9090": Address and port to bind listener to
  -remote="": Address and port of remote graphite instance [required]
```

## Build and Test

Build:
```SHELL
make
```

Test:
```SHELL
make test
make bench
```

### TODO
- Encryption
- Better error handling when the backend connections drop
- More tests (only the parsing and filtering is tested/benched at this point)

# License

MIT - Samuel Giles 2015
