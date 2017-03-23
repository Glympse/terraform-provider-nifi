# Changelog

## Known Limitations

- Only Process Groups, Processors and Connection resources are supported. 
- Parent group id can't be changed (on any of the resources).
- Processor and connection update and delete operations cannot be parallelized. 
  Explicit locking is used to prevent those from being run concurrently.   
  See [nifi/client.go](nifi/client.go) for details. 
- Connection data is being dropped prior to connection removal.
  Plugin does so in order to automate flow transformations. 
  Production applications should consider ensuring that connections that are subject to removal are properly purged
  prior to running `terraform apply`.  

## 0.2.0

- All processors are kept in `RUNNING` state (unconditionally).
- Proper processor state transitions are performed when processors and connections are updated.  

## 0.1.0

- Support for Process Group, Processor and Connection resources.
