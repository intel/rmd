# RMD User Guide:

## Pre-requires

Please read the [requirements][requirements].

[requirements]: https://github.com/intel/rmd/blob/master/docs/PreRequirment.md

## RMD processes

For security considerations, RMD daemon runs the RESTAPI service (a spawned
child) as a normal unix user 'rmd' process and performs privileged operations
(PAM authentication or resctrl) as 'root' unix user.

Ensure the following libraries are installed on your target system to use PAM
Debian
```
sudo apt-get install openssl libpam0g-dev db-util
```
Redhat
```
sudo dnf install openssl pam-devel db4-utils
```

Clients are validated against host unix credentials or credentials setup in a
Berkeley database file.
To setup users in Berkeley database use the bash script:

```
./setup_rmd_users.sh
```
Note: Only root user can setup or access users in Berkeley database.

Launch RMD by supplying configuration directory using the command

```
sudo ./rmd --conf-dir ./etc/rmd
```

`RMD can be launched in debug mode that exposes RESTAPI service on HTTP.
By default it is launched on port 8081. (http://127.0.0.1:8081)`

## Query cache information on the host

```
curl -i http://127.0.0.1:8081/v1/cache/
curl -i http://127.0.0.1:8081/v1/cache/l3
curl -i http://127.0.0.1:8081/v1/cache/l3/0
```

## Query pre-defined policy in RMD

```
curl http://127.0.0.1:8081/v1/policy
```

The backend for policy is a yaml file: /etc/rmd/policy.yaml, it pre-defines
some `policy` on specific intel platform. This file can be changed.


## Configuration file example

```
[OSGroup] # OSGroup is mandatory
cacheways = 1
cpuset = "0-1"

[InfraGroup] # InfraGroup is optional
cacheways = 19
cpuset = "2-3,24-25"
# arrary or comma-separated values? RMD supports array instead of CSV.
tasks = ["ovs*"] # Just support Wildcards. Do we need to support RE?

[CachePool]
shrink = false # wether allow to shrink cache ways in best effort pool
max_allowed_shared = 10 # max allowed workload in shared pool, default is 10
guarantee = 10
besteffort = 7
shared = 2
```

- OSGroup: cache ways reserved for operate system usage.
- InfraGroup: infrastructure tasks group, user can specify task binary name
              the cache ways will be shared with other worklaod

Besides, define cache pool, cache allocation will happened in the pools.

- shrink: whether shrink cache ways which already allocated to workload in
          besteffort pool.
- max_allowed_shared: max allowed workloads in shared cache pool
- guarantee: allocate cache for workload max_cache == min_cache > 0
- besteffort: allocate cache for workload max_cache > min_cache > 0
- shared: allocate cache for workload max_cache == min_cache = 0

On a host which support max 20 cache ways, for this configuration file,
we will have follow cache bit mask layout:

```

OSGroup:    0000 0000 0000 0000 0001
InfraGroup: 1111 1111 1111 1111 1110
```
Available CBM in Cache Pools initially:
```
guarantee:  0000 0000 0111 1111 1110
besteffort: 0011 1111 1000 0000 0000
shared:     1100 0000 0000 0000 0000
```

## Create a workload

A workload could be a running task(s) or some cpus which want to allocate
cache for them.

The task(s)/cpus should be valided, RMD will fail your request they are
invalid.

Besides you need to specify what policy of the workload will be usage.

The post body could contains:

```
{
    "task_ids": A validate task id list
    "core_ids": cpu core list, for the topology, check cache information
    "policy": pre-defined policy in RMD
    "max_cache": maximum cache ways which can be benefited
    "min_cache": minmum cache ways which can be benefited
}
```

You can not neither specify policy and max_cache/min_cache at same time, that
is ambiguous to RMD.

An example:

1) Create a workload with gold policy.

```
curl -H "Content-Type: application/json" --request POST --data \
         '{"task_ids":["78377"], "policy": "gold"}' \
         http://127.0.0.1:8888/v1/workloads
```

2) Create workload with max_cache, min_cache.

```
curl -H "Content-Type: application/json" --request POST --data \
         '{"task_ids":["14988"], "max_cache": 4, "min_cache": 4}' \
         http://127.0.0.1:8888/v1/workloads
```

Admin can change and add new policies by editing an toml/yaml file which is
pointed in the configuration file.

```
policypath = "etc/rmd/policy.toml"
```

## Hospitality score API usage:

Hospitality score API will give a score for scheduling workload on a host for
cache allocation request.

Admin can ether give the max_cache/min_cache or a policy to query if the
hospitality score.

The score will be calculate as following:

| request | hospitality score | cahce pool |
| :-----: | :---------------: | :--------: |
| max_cache == min_cache > 0 | `[0 \| 100]` | Guarantee |
| max_cache == min_cache == 0 | `[0 \| 100]` | Shared |
| max_cache > min_cache > 0 |  `[0 , 100]` | Besteffort |


To get hospitality score:

```
curl -H "Content-Type: application/json" --request POST --data \
         '{"max_cache": 2, "min_cache": 2}' \
         http://127.0.0.1:8888/v1/hospitality
{
    "score": {
        "l3": {
            "0": 100,
            "1": 100
        }
    }
}
```

## Access RMD by unix socket:

Access RMD by unit socket if it is enabled.

Requires curl >= v7.40.0
```
sudo curl --unix-socket /your/socket/path http:/your/resource/url
```

## Access RMD by TLS:

Access RMD by TLS if it is enabled.

Need to config tlsport, certpath, clientcapath, clientauth options in
configure file.

Using TLS and managing a CA is an advanced topic. It is not the scope of RMD.
RMD just pre-define server certs for testing.

Please do not use them in product environment.
User can generate certs by themselves.

If you want to get cache info, your can run this command:
```
curl https://hostname:port/v1/cache --cert etc/rmd/cert/client/cert.pem \
         --key etc/rmd/cert/client/key.pem \
         --cacert  etc/rmd/cert/client/ca.pem
```

[1]: http://www.intel.com/content/www/us/en/processors/architectures-software-developer-manuals.html
