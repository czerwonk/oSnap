# oSnap
[![Build Status](https://travis-ci.org/czerwonk/oSnap.svg)](https://travis-ci.org/czerwonk/oSnap)
[![Go Report Card](https://goreportcard.com/badge/github.com/czerwonk/osnap)](https://goreportcard.com/report/github.com/czerwonk/osnap)

Create virtual machine snapshots (using the oVirt API) with one single command

## Install
```
go get -u github.com/czerwonk/osnap
```

## Configuration
oSnap is configured by a YAML based config file:

```yaml
api:
  url: https://my-ovirt.net
  user: my-osnap-user
  password: my-pass

cluster: my-cluster
keep: 3

includes:
  - web.*
  - app.*
 
excludes:
  - db.*
  - temp.*
```

## License
(c) Daniel Czerwonk, 2017. Licensed under [MIT](LICENSE) license.

## oVirt
see https://www.ovirt.org/
