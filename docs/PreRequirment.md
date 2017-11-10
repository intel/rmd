# Pre-requires

## Hardware

RMD requires specific hardware support. It requires the host to have
[Intel(R) Resource Director Technology][1] features, including below:

[1]: https://www.intel.com/content/www/us/en/architecture-and-technology/resource-director-technology.html

- Intel(R) Xeon(R) processor E5 v3
- Intel(R) Xeon(R) processor D
- Intel(R) Xeon(R) processor E3 v4
- Intel(R) Xeon(R) processor E5 v4
- Intel(R) Xeon(R) Scalable Processors (6)
- Intel(R) Atom(R) processor for Server C3000

## Software

RMD depends on linux `resctrl` sysfs, for upstream linux kernel version,
it should be higher than 4.10, or any linux distro which has enabled `resctrl`
interface.

To check if your host supports `resctrl` or not, check the out put of this
command line:

```
cat /proc/filesystems  | grep resctrl
```
