# Forward Middleware Plugin

A Traefik middleware plugin that automatically adds the `X-Forwarded-For` header to HTTP requests using the client's remote IP address.

[![Build Status](https://github.com/traefik/plugindemo/workflows/Main/badge.svg?branch=master)](https://github.com/traefik/plugindemo/actions)

## What it does

This plugin extracts the client's IP address from the request's `RemoteAddr` field and sets it as the `X-Forwarded-For` header if one doesn't already exist. This is useful for:

- Preserving client IP information when requests pass through proxies
- Enabling downstream services to identify the original client IP
- Meeting compliance requirements that need client IP logging
- **Solving mixed traffic scenarios** (see below)

The plugin safely handles malformed remote addresses by failing gracefully and passing the request through unchanged.

## Mixed Traffic Scenarios

A common problem occurs when you have both direct requests to Traefik and requests coming through a Layer 7 load balancer. In this scenario:

- **Direct requests**: Don't have `X-Forwarded-For` headers, so IP allowlists work with the actual client IP
- **Load balancer requests**: Have `X-Forwarded-For` headers, but IP allowlists see the load balancer's IP instead of the client IP

This creates an inconsistent situation where you can't reliably use Traefik's IP allowlist middleware with `ipStrategy` settings.

### The Solution

This plugin solves the problem by ensuring **all requests have consistent `X-Forwarded-For` headers**:

1. Requests through load balancer: Already have `X-Forwarded-For` → Plugin does nothing
2. Direct requests: Missing `X-Forwarded-For` → Plugin adds it using the client IP

Now you can reliably use Traefik's IP allowlist with `ipStrategy`:

```yaml
middlewares:
  forward-headers:
    plugin:
      forward-middleware:
        enabled: true
        
  ip-allowlist:
    ipAllowList:
      sourceRange:
        - "192.168.1.0/24"
        - "10.0.0.0/8"
      ipStrategy:
        depth: 1  # Use the rightmost IP in X-Forwarded-For
        # OR
        excludedIPs:  # Exclude known load balancer IPs, you can also use CIDR ranges
          - "192.168.100.1"
          - "192.168.100.2"
          - "192.168.100.0/16"

http:
  routers:
    my-router:
      middlewares:
        - forward-headers  # Apply FIRST
        - ip-allowlist     # Apply SECOND
```

**Important**: The forward middleware must be applied **before** the IP allowlist middleware in the chain.


### Configuration

```yaml
# Static configuration

experimental:
  plugins:
    forward-middleware:
      moduleName: github.com/hiasr/forwardmiddleware
      version: v0.1.0
```

Here is an example of a file provider dynamic configuration (given here in YAML), where the interesting part is the `http.middlewares` section:

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - forward-headers

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    forward-headers:
      plugin:
        forward-middleware:
          enabled: true
```

