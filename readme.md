# Terraform NiFi Provider

The NiFi provider is used to interact with NiFi cluster.
The provider needs to be configured with the proper host before it can be used.

Use the navigation to the left to read about the available resources.

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
