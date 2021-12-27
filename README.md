# Eclipso - High performance DNS daemon

```
┌─┐┌─┐┬  ┬┌─┐┌─┐┌─┐
├┤ │  │  │├─┘└─┐│ │
└─┘└─┘┴─┘┴┴  └─┘└─┘
v1.0.1
```

# Project objectives

* High performance - Speed wins
* Lightweight - Bloat and large dependencies are not welcome
* Container first - Built to be run as a container with support for Amazon ECS/EKS and Kubernetes
* Ease of use - K.I.S.S

# Installation

```
git clone https://github.com/benduncan/eclipso
cd eclipso
make build
```

## Running eclipso

```
./bin/eclipso
```

## Zone file storage engines

* Local filesystem (default)
* S3 bucket (Env variable configuration required)
* DynamoDB (Roadmap)

## Zone file resync

Given the performance requirements, Eclipso maintains a local hashmap containing the zone file database and does not introduce an external database or additional network latency to DNS domain requests.

Eclipso will monitor the following to live reload zone file changes:

* Local filesystem - Automatically live reload zone files based on filesystem events (add, modified & removed files)
* S3 Bucket Sync - Periodic scheduler to compare the remote S3 bucket to the local zone database
* S3 Bucket events - Roadmap, receive S3 events via SQS/SNS to live changes in the bucket
* DynamoDB Stream - Roadmap, receive a live stream from DynamoDB to sync local state

## Environment variables

Configuration of Eclipso is made via environment variables.

### Local directory config

```
ZONE_DIR="config/domains"
```

### S3 configuration

To read configuration files from a specified S3 bucket

```
AWS_ACCESS_KEY="XXX"
AWS_SECRET_ACCESS_KEY="YYY"
ZONE_DIR="s3://bucket-name"
AWS_REGION="us-west-1" 
```

Optional parameters

```
S3_SYNC_RETRY="120"
```

Specify the S3 bucket sync period in seconds (default 60).

### Global config

```
PORT=5353
```

Run Eclpiso on the specified (UDP) port 5353. Default 53.

```
ECLIPSO_LOG_IGNORE=1
```

Disable verbose logging

```
ECLIPSO_LOG_DEBUG=1
```

Enable additional debug logs

# Zone configuration file

Example configuration file

```
# Domain configuration file in TOML format.
version = 1.0

# Domain settings
[domain]
domain = "hello_a.net"
created = 2021-05-27T07:32:00Z
modified = 2022-05-27T07:32:00Z
verified = true
active = true
ownerid = 10

# Default settings if not defined in each [[records]]
[defaults]
ttl = 3600
type = 1
class = 1

# Domain entry, one entry per record
# A record for root domain, hello_a.net > 203.100.1.1
[[records]]
domain = ""
address = "203.100.1.1"

# A record for www.hello_a.net > 203.100.1.1
[[records]]
domain = "www."
address = "203.100.1.1"

# A record for host-1.hello_a.net > 203.100.1.2
[[records]]
domain = "host-1."
address = "203.100.1.2"

# A record for host-1.hello_a.net > 203.100.1.3
[[records]]
domain = "host-2."
address = "203.100.1.3"

# MX record for hello_a.net, with MX preference set. 
[[records]]
domain = ""
type = 15
preference = 10
address = "host-1.hello_a.net."

# MX record for hello_a.net, with MX preference set. 
[[records]]
domain = ""
type = 15
preference = 20
address = "host-2.hello_a.net."

# TXT record to specify SPF record
[[records]]
domain = ""
type = 16
address = "v=spf1 ip:203.100.1.2 ip:203.100.1.3 mx a -all"

# Sample TXT record for google-site-verification
[[records]]
domain = ""
type = 16
address = "google-site-verification=3a13e1788d7a1c3b4602afc083e855de"
```

To run Eclipso with the sample domain, append the configuration file to `./config/domains/hello_a.net.toml`

Run Eclpiso:

`ZONE_DIR="./config/domains" ./bin/eclipso`

Query the local instance to validate:

```
$ dig @127.0.0.1 hello_a.net txt

;; QUESTION SECTION:
;hello_a.net.			IN	TXT

;; ANSWER SECTION:
hello_a.net.		3600	IN	TXT	"v=spf1 ip:203.100.1.2 ip:203.100.1.3 mx a -all"
hello_a.net.		3600	IN	TXT	"google-site-verification=3a13e1788d7a1c3b4602afc083e855de"

$ dig @127.0.0.1 hello_a.net mx

; <<>> DiG 9.10.6 <<>> @127.0.0.1 hello_a.net mx
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 43492
;; flags: qr aa rd; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;hello_a.net.			IN	MX

;; ANSWER SECTION:
hello_a.net.		3600	IN	MX	10 host-1.hello_a.net.
hello_a.net.		3600	IN	MX	20 host-2.hello_a.net.
```

## Run within docker

Using docker-compose (S3 bucket):

```
AWS_ACCESS_KEY="X" AWS_SECRET_ACCESS_KEY="Y" ZONE_DIR="s3://my-bucket" AWS_REGION="us-west-1" docker-compose up -d
```

Standalone docker (filesystem method):

```
docker run --mount src=~/eclipso/config/domains,target=/config/domains,type=bind -e ZONE_DIR="/config/domains" -p 53:53/udp calacode/eclipso-dns
```

## Run as a standalone daemon

```
ZONE_DIR=./config/domains ./bin/eclipso
```

# High-level Roadmap

Version 1.0.X:

- [x] DNS server framework
- [x] Filesystem support for zone files
- [x] Filesystem live-reload events
- [x] S3 Bucket support for zone files
- [x] S3 resync period for live-reload
- [x] Implement a hashmap for the local zone database
- [x] Docker support

Immediate roadmap:

- [ ] Support additional zone types (SRV)
- [ ] DynamoDB support
- [ ] DynamoDB stream support
- [ ] Improve benchmarking
- [ ] Improve CLI flag support and env usage

General roadmap:

- [ ] Optimise local hashmap use Net.IP (v4/v6) types
- [ ] Additional API daemon to add/delete/update zone files
- [ ] Health-check support, change return address if end-point not answering
- [ ] DNS resolver option - Raw lookups via root-servers to act as a DNS forwarder

# Benchmarking

Install `benchstat` to compare multiple benchmarks for a more accurate reading

```
go get golang.org/x/perf/cmd/benchstat
```

Next, run the benchmark:

```
make bench
```

This will simulate 24 domains (hello_a.net ... hello_z.net) with ~255 sub-domains (host-1 ... host-255) with sample TXT records. The benchmark script will query the local instance each record for benchmarking purposes.

```
ECLIPSO_LOG_IGNORE=1 go test -bench=. ./pkg/backend -count 5 -benchmem | tee benchmark.out
goos: darwin
goarch: amd64
pkg: github.com/benduncan/eclipso/pkg/backend
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkDNSQueryA-8     	    8972	    175261 ns/op	    3088 B/op	      58 allocs/op
BenchmarkDNSQueryA-8     	    8598	    163153 ns/op	    3087 B/op	      58 allocs/op
BenchmarkDNSQueryA-8     	    8149	    162703 ns/op	    3089 B/op	      58 allocs/op
BenchmarkDNSQueryA-8     	    6561	    155536 ns/op	    3088 B/op	      58 allocs/op
BenchmarkDNSQueryA-8     	    8355	    140921 ns/op	    3088 B/op	      58 allocs/op
BenchmarkDNSQueryTXT-8   	    7900	    145299 ns/op	    3682 B/op	      68 allocs/op
BenchmarkDNSQueryTXT-8   	    8028	    168382 ns/op	    3681 B/op	      68 allocs/op
BenchmarkDNSQueryTXT-8   	    7179	    154651 ns/op	    3681 B/op	      68 allocs/op
BenchmarkDNSQueryTXT-8   	    7720	    204199 ns/op	    3680 B/op	      68 allocs/op
BenchmarkDNSQueryTXT-8   	    8929	    186801 ns/op	    3680 B/op	      68 allocs/op
BenchmarkDNSQueryMX-8    	    7741	    174948 ns/op	    4049 B/op	      76 allocs/op
BenchmarkDNSQueryMX-8    	    9464	    152655 ns/op	    4050 B/op	      76 allocs/op
BenchmarkDNSQueryMX-8    	    7480	    153643 ns/op	    4049 B/op	      76 allocs/op
BenchmarkDNSQueryMX-8    	    6700	    181419 ns/op	    4049 B/op	      76 allocs/op
BenchmarkDNSQueryMX-8    	    8418	    148584 ns/op	    4051 B/op	      76 allocs/op
PASS
ok  	github.com/benduncan/eclipso/pkg/backend	31.133s
benchstat benchmark.out
name           time/op
DNSQueryA-8     160µs ±12%
DNSQueryTXT-8   172µs ±19%
DNSQueryMX-8    162µs ±12%

name           alloc/op
DNSQueryA-8    3.09kB ± 0%
DNSQueryTXT-8  3.68kB ± 0%
DNSQueryMX-8   4.05kB ± 0%

name           allocs/op
DNSQueryA-8      58.0 ± 0%
DNSQueryTXT-8    68.0 ± 0%
DNSQueryMX-8     76.0 ± 0%
```
