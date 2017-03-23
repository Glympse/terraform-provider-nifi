# Changelog

## Known Limitations

- Only Process Groups, Processors and Connection resources are supported. 
- Parent group id can't be changed (on any of the resources).

## 0.2.0

- All processors are kept in `RUNNING` state (unconditionally).
- Proper processor state transitions are performed when processors and connections are updated.  

## 0.1.0

- Support for Process Group, Processor and Connection resources.
