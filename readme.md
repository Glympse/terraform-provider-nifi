# Terraform NiFi Provider

The NiFi provider is used to interact with NiFi cluster.
The provider needs to be configured with the proper host before it can be used.

Use the navigation to the left to read about the available resources.

## Plugin Requirements

- Terraform 0.9

## NiFi Version Compatibility

Plugin Version | Supported NiFi API Version
---|---
0.1+ | 1.1+

## Known Limitations

- Only Process Groups, Processors and Connection resources are supported. 
- Parent group id can't be changed (on any of the resources).
- Resources are never started or stopped.
  
## References

- NiFi API Documentation
  https://nifi.apache.org/docs/nifi-docs/rest-api/
