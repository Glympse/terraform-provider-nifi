# Syntax

## Example Usage

```
# Configure the NiFi provider
provider "nifi" {
  host        = "localhost:8080"
  api_path    = "nifi-api"
  admin_cert  = "certs/nifi.crt"
  admin_key   = "certs/nifi.key"
  http_scheme = "http"
}
```

## Argument Reference

The following arguments are supported:

Argument         | Required | Description
-----------------|----------|------------
**host**         | Yes      | NiFi host including port, e.g. `localhost:8080`.
**api_path**     | No       | API path prefix, e.g. `nifi-api`. Defaults to that.
**admin_cert**   | No       | Path to certificate used to access admin. Provider will use HTTPS only if this is specified.
**admin_key**    | No       | Path to certificate's key, required if `admin_cert` is specified.
**http_scheme**  | No       | Force a HTTP scheme. Useful if NiFi does not handle SSL termination. Defaults to `http`, unless `admin_cert` and `admin_key` are set, in which case `https` is used.
