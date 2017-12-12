
<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Resource Management Daemon #

----------

Resource Management Daemon (RMD) is a system daemon running on generic Linux platforms. The purpose of this daemon is to provide a central uniform interface portal for hardware resource management tasks on x86 platforms.

----------
## Overview ##
RMD manages [Intel RDT](https://www.intel.com/content/www/us/en/architecture-and-technology/resource-director-technology.html) resources as the first step. Specifically in the current release, Cache Allocation Technology (CAT) is supported. CAT hardware feature is exposed to the software by a number of Model Specific Registers (MSR). It is supported by several software layers (e.g., libpqos and resctrl file system). The advantages of RMD are:


* **User friendly API**: Most (if not all) of the alternative ways to use RDT resources include manipulating bit masks，whereas RMD offers a user friendly RESTFul API that end users just need to specify the amount of the desired resources and some other attributes. RMD will convert that quantity into corresponding bit masks correctly and automatically.
* **System level awareness**: One system may (and quite possible in a hyper-convergent deployment) host several software entities like OpenStack, Kubernates, Ceph and so on. Each of these software entities may have their built-in support for RDT resources but they may not have a system level view of all the competitors of RDT resources and thus lacks of coordination. Through RMD, these software entities can collaborate in resource consumption. RMD can be a system level resource orchestrator.
* **Built-in intelligence**: **Though not supported yet**, in RMD road map, Machine Learning is one of the attractive incoming features which will provide intelligence to auto adjust resource usage according to user pre-defined policies and the pressure of resource contention. 


### Cache Pools/Groups ###
RMD divides the system L3 cache into the following groups or pools. Each task of a RMD enabled system falls into one of the groups explicitly or implicitly. Workloads are used to describe a group of tasks of the same cache attributes.

* **OS group**: This is the default cache group that any newly spawned task on the system is put into if not specified otherwise. Tasks in this group all shares the cache ways allocated to this group but does not share/overlap with cache ways in other groups.
* **Infra group**: Infrastructure group. Tasks allocating cache ways from this group share cache ways with all of the other groups **except** OS group. This group is intended for the infrastructure software that provides common facilitation to all of the workloads. An example would be the virtual switch software that connects to all the virtual machines in the system.
* **Guaranteed group**: Workloads allocating cache ways from this group have their guaranteed amount of desired cache ways. Cache ways in this group are dedicated to their associated workloads, not shared with others except the infra group.
* **Best effort group**: Workloads allocating cache ways from this group have their minimal amount of desired cache ways guaranteed but can burst to their maximum amount of desired cache ways whenever possible. Cache ways in this group are also dedicated to their associated workloads, not shared with others except the infra group.
* **Shared group**: Workloads allocating cache ways from the shared group shares the whole amount of cache ways assigned to the group.

The amount of cache ways for each of the above groups are configurable in the RMD configuration file. Below diagram gives an example of a system of 11 cache ways. 

![RMD Groups](https://github.com/intel/rmd/tree/master/docs/pic/rmd_pools.png)

### Cache Specification ###
Please refer to the [API documentation](docs/api/v1/swagger.yaml) for a comprehensive description of RMD APIs. Here is a brief depiction of how to assign workloads to different aforementioned cache pools.

OS group is the default group, so if no one explicitly moves a task or workload to other group, then it stays in the OS group.

Tasks in the infra group are pre-configured in the configuration file. No API is provided to assign a task to the infra group dynamically.

End users make their cache requirements by specifying two values ("max\_cache" and "min\_cache") associated with the workload:

* max\_cache == min\_cache > 0 &nbsp;&nbsp;&nbsp;&nbsp; ==> guaranteed group
* max\_cache >  min\_cache > 0 &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; ==> best effort group
* max\_cache == min\_cache == 0&nbsp;&nbsp;&nbsp; ==> shared group

## Architecture ##

From a logical point of view, there are several components of RMD:

* HTTPS server -- provides mutual (client and server) authentication and traffic encryption
* RESTFul API provider -- accepts and sanitizes user requirements 
* Policy engine -- decides whether to enforce or reject user requirement based on system resource status
* Resctrl filesystem interface -- interacts with kernel resctrl interface to enforce user requirements

![RMD logical view](https://github.com/intel/rmd/tree/master/docs/pic/rmd_logical_view.png)

From a physical point of view, RMD is composed of two processes -- the front-end and the back-end. The splitting of RMD into two processes is of security concerns. The front-end process which conducts most of the jobs runs as a normal user (least privilege). Whereas the back-end process runs as a privileged user because it has to do modifications to the resctrl file system. The back-end process is deliberately kept as small/simple as possible. Only add logic to the back-end when there is definitely a need to lift privilege. The front-end and back-end communicates via an anonymous pipe.

For more information on the design and architecture, please refer to the [developers guide](docs/developers_guide.md)

## API Introduction ##
Please refer to the [API documentation](docs/api/v1/swagger.yaml) for a comprehensive description of RMD APIs. This section provides the introduction and rationale of the API entry points.

### "/cache" entry point ###
This entry point and its sub-categories are to get system cache information. so only "GET" method is accepted by this entry point.

### "/workloads" entry point ###
Through the "/workloads" entry point you can specify a workload by CPU IDs and/or task IDs. And specify the workload's demand of caches in one of two ways. The first way is to specify the "max\_cache/min\_cache" values explicitly as aforementioned. The second way is to associate the workload with one of the pre-defined "policies" (see below "/policy" entry point). The pre-defined policies have pre-defined "max\_cache/min\_cache" values that they are translated into.

### "/hospitality" entry point ###
The reason behind this "/hospitality" entry point is that there are often the needs to know how well a host can do to fulfill a certain cache allocation requirement. This requirement usually comes from scheduling in a large cluster deployment. So the notion of "hospitality score" is introduced.

Why can't the available cache amount do the job? Currently the last level cache in Intel platforms can only be allocated contiguously. So the totally amount of available last level cache won't help due to segmentation issues.

The hospitality score is calculated differently for workloads of different cache groups. (In below explanation 'value' means the largest available contiguous cache ways in the corresponding group)

* guaranteed group:<br>
  `if value > max_cache then return 100 else return 0`
* best effort group:<br>
  `if value > max_cache then return 100`<br>
  `if min_cache < value < max_cache then return (value/max)*100`<br>
  `if value < min_cache then return 0`
* shared group:<br>
  `return 100`


### "/policy" entry point ###
The "/policy" entry point contains the pre-defined recommended cache usage values for the specific platform that this RMD instance is running. Though completely configurable, the default policies are defined as "Gold/Sliver/Bronze" to classify different service levels. API user can get policies and associate workloads with one of the policies.

## Refereneces ##
[Deployment guide](docs/deployment_guide.md)

[API Documentation](docs/api/v1/swagger.yaml)

[Users guide](docs/UserGuide.md_)

[Developers guide](docs/developers_guide.md)
