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

RMD cache module depends on linux `resctrl` sysfs. For `resctrl` support upstream linux kernel version
should be higher than 4.10 or any linux distro which has enabled `resctrl` interface is needed.

To check if your host supports `resctrl` or not, check the out put of this
command line:

```
cat /proc/filesystems  | grep resctrl
```

Optional external `pstate` module requires `intel_pstate` or `acpi` CPU scaling driver to monitor and change the CPU cores frequencies.

To check which driver is used on your host run following command in your Linux shell:

```
cat /sys/devices/system/cpu/cpufreq/policy0/scaling_driver
```

This module is a dynamically loadable external plugin and is delivered separately.