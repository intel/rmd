# Resource Management Daemon changelog

## Version 0.2

### Bugfixing and design change

List of changes
- project changed to go module (Go >= 1.11 required)
- code restructured for modular architecture
- updated travis and rpm packaging files
- refactored workload and policy related modules
- introduced modular interface for resource plugins
- updated unit and functional tests
- implemented PAM library wrapper
- RMD TLS connection forced to use AES-256/SHA-384 cipher suite
- cleaned code and fixed bugs

## Version 0.1

### Initial release
