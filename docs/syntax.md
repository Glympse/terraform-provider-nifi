# Syntax

## Example Usage

```
# Configure the NiFi provider
provider "nifi" {
  host = "localhost:8080"
  api_path = "nifi-api"
}
```

## Argument Reference

The following arguments are supported:

Argument | Required | Description
---|---|---
**api_key** | Yes | NiFi host (including port). e.g. `localhost:8080`
**api_path** | No | API path prefix. e.g. `nifi-api`
