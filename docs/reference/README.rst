# Reference Docs

This directory contains some reference dos for RDT usage.

## CAT

There are 2 ways to use CAT feature [0]_:

1. Associate logical thread with logical CLOSS

  a): Determine whether the CPU supports CAT via the CPUID instruction. As described in the Intel Software Developer’s Manual, CPUID leaf 0x10 provides detailed information on the capabilities of the CAT feature.

  b): Configure the class of service (CLOS) to define the amount of resources (cache space) available via MSRs.

  c): Associate each logical thread with an available logical CLOS.

  d): As the OS/VMM swaps a thread or VCPU onto a core, update the CLOS on the core via the IA32_PQR_ASSOC MSR, which ensures that the resource usage is controlled via the bitmasks configured in step 2.

In this way, we need OS/VMM enabling, this part of work has been done by Linux kernel 4.10. see `linux-kernel-user-interface.txt` for details.

We can leverage this OS/VMM interface and no worry about CLOS -> hard core association. OS scheduler should in charge of migrate CLOS to core.

Linux kernel provides /sys/fs/resctrl interface to operate on each cache resource [1]_


2. Pin CLOS to hardware thread(Cpu Core)

`An alternate usage method is possible for non-enabled operating systems and VMMs where CLOS are pinned to hardware threads, then software threads are pinned to hardware threads; however OS/VMM enabling is recommended wherever possible in order to avoid the need to pin apps.`

In this way, we need to pin our APP/VM threads on some specified cores, then associate CLOS to cores, this is what libpqos does now.
See `libpqos-user-interface.txt` for how to use CAT features [2]_.

.. [0] cat usage: https://software.intel.com/en-us/articles/cache-allocation-technology-usage-models
.. [1] intel rdt ui: https://chromium.googlesource.com/external/github.com/altera-opensource/linux-socfpga/+/refs/heads/master/Documentation/x86/intel_rdt_ui.txt
.. [2] intel-cmt-cat: https://github.com/01org/intel-cmt-cat/tree/master/rdtset
