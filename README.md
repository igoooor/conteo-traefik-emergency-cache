Note: This repository is public and can be found at: https://github.com/igoooor/conteo-traefik-emergency-cache
Traefik pod will git pull from that public repo. The bitbucket copy is just there for reference 

# Conteo Traefik Cache

This is a fork from https://github.com/traefik/plugin-simplecache

Simple cache plugin middleware caches responses on disk.

## Configuration

To configure this plugin you should add its configuration to the Traefik dynamic configuration as explained [here](https://docs.traefik.io/getting-started/configuration-overview/#the-dynamic-configuration).
The following snippet shows how to configure this plugin with the File provider in TOML and YAML: 

Static:

```toml
[pilot]
  token="xxx"

[experimental.plugins.cache]
  modulename = "github.com/igoooor/conteo-traefik-emergency-cache"
  version = "v0.0.1"
```

Dynamic:

```toml
[http.middlewares]
  [http.middlewares.my-cache.plugin.cache]
    path = "/some/path/to/cache/dir"
```

```yaml
http:
  middlewares:
   my-cache:
      plugin:
        cache:
          path: /some/path/to/cache/dir
```

### Options

#### Path (`path`)

The base path that files will be created under. This must be a valid existing
filesystem path.

#### Max Expiry (`maxExpiry`)

*Default: 300*

The maximum number of seconds a response can be cached for. The 
actual cache time will always be lower or equal to this.

#### Cleanup (`cleanup`)

*Default: 600*

The number of seconds to wait between cache cleanup runs.
	
#### Add Status Header (`addStatusHeader`)

*Default: true*

This determines if the cache status header `Cache-Status` will be added to the
response headers. This header can have the value `hit`, `miss` or `error`.
