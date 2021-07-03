# eclipso

## Basic Roadmap

* Use distributed k/v system in Go, native
* Support ipv4
* MX, A, CNAME support
* Master multi-multi mode.
* Simple API to add/delete/update records
    * eclispo auth (id/key)
    * eclipso add domain.com A 1.1.1.1
    * eclipso delete domain.com
    * elcipso update domain.com A 2.2.2.2

## Future roadmap
* Support ipv6
* Raw lookups via root-servers, no need for local NS