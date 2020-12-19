# `xbps-cache`, a caching proxy for Void Linux packages

## Config file `/etc/xbps-cache.conf`

```toml
# Port and IP to listen on
# LocalEndpoint = ":8081"

# Path for downloaded files. Must exist and be writable by cache user
StoreDir = "/var/cache/xbps-cache"

# Path for logs. Must exist and be writable by cache user
LogDir = "/var/log/xbps-cache"

# Uplink server to query
#UplinkURL = "https://alpha.de.repo.voidlinux.org"
```

## Service file `/etc/sv/xbps-cache/run`

```sh
#!/bin/sh

export USERNAME=xbps-cache
exec chpst -u $USERNAME /usr/bin/xbps-cache
```

## Script to override repository definitions

```sh
#!/bin/sh

# Enter URL of configured cache
LOCAL=http://192.168.1.60:8081

# Original URL to replace
ORIG=https://alpha.de.repo.voidlinux.org

for REPO in /usr/share/xbps.d/*-repository-*.conf ; do
    BASE=$(basename $REPO)
    sed "s#${ORIG}#${LOCAL}#" $REPO > /etc/xbps.d/$BASE
done
```
