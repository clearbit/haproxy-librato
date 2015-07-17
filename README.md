# haproxy-librato
Submit haproxy stats to librato

# Usage

```
HAPROXY_URL=http://haproxy.acme.com/stats;csv \
LIBRATO_USER=user@example.com \
LIBRATO_TOKEN=sekret \
LIBRATO_SOURCE=internal-balancer-1 \
./haproxy-librato
```

`haproxy-librato` will submit stats to librato every 30 seconds.
