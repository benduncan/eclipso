# Eclipso - High performance DNS daemon

```
┌─┐┌─┐┬  ┬┌─┐┌─┐┌─┐
├┤ │  │  │├─┘└─┐│ │
└─┘└─┘┴─┘┴┴  └─┘└─┘
```

# Project objectives

* High performance - Speed wins
* Lightweight - Bloat and large dependency chains not welcome
* Container first - Built to be run as a container

# Installation

```
make build
```

## Benchmarking

```
make bench
```

Will spin up 100,000 local domains and randomly query a subset of local hosted domains within a 60 sec period to benchmark throughput.

## Run within docker



## Run as a standalone daemon

## Configuration

## Storage methods

* Local filesystem
* S3 bucket
* S3 bucket (with SQS notofications, pending)
* DyanmoDB (pending)



## Basic Roadmap

X * Use packages/modules
* Use S3 config to start - or shared config file between nodes?
* Use distributed k/v system in Go, native
* Support ipv4
* MX, A, CNAME support
* Master multi-multi mode.
* Simple API to add/delete/update records
    * eclispo auth (id/key)
    * eclipso add domain.com A 1.1.1.1
    * eclipso delete domain.com
    * elcipso update domain.com A 2.2.2.2
* healthcheck support, like R53, change IP if end-point not answering
* add preference to MX/config settings
* add missing settings (like RFC zone file) - SOA, nameserver, hostmaster, expiry, serial, etc
* load multiple domains, scan directory
* file changes, reload settings. If on disk or S3 trigger
* MX record preferences in config setting

## Test cases

* config file load
    * bad domains, whitespace, missing . checks
* test DNS response, multi A records for example



## Future roadmap
* Support ipv6
* Raw lookups via root-servers, no need for local NS