[![Build Status](https://drone.whyrl.fr/api/badges/Whyrl35/prometheus-saltstack-exporter/status.svg)](https://drone.whyrl.fr/Whyrl35/prometheus-saltstack-exporter)  [![Go Report Card](https://goreportcard.com/badge/github.com/Whyrl35/prometheus-saltstack-exporter)](https://goreportcard.com/report/github.com/Whyrl35/prometheus-saltstack-exporter)  [![GitHub release (latest by date)](https://img.shields.io/github/v/release/Whyrl35/prometheus-saltstack-exporter)](https://github.com/Whyrl35/prometheus-saltstack-exporter/releases/latest)

# prometheus-saltstack-exporter

Export [Saltstack](https://saltproject.io/) metrics to [Prometheus](https://prometheus.io/)

- [prometheus-saltstack-exporter](#prometheus-saltstack-exporter)
  - [Metrics](#metrics)
  - [Preparing saltstack API](#preparing-saltstack-api)
    - [create at least one external auth user or group](#create-at-least-one-external-auth-user-or-group)
    - [rest\_cherrypy configuration](#rest_cherrypy-configuration)
    - [testing API](#testing-api)
      - [login](#login)
  - [Usage](#usage)
  - [Installing](#installing)
    - [Debian family](#debian-family)
    - [RedHat family](#redhat-family)
    - [Others](#others)
- [Development](#development)
  - [Install go environment](#install-go-environment)
  - [Makefile](#makefile)
  - [Testing](#testing)


## Metrics

Masters metrics

```
# HELP saltstack_master_up Master in up(1) or down(0)
# TYPE saltstack_master_up gauge
saltstack_master_up{master="saltmaster.local.domain"} 1
```

Minions metrics

```
# HELP saltstack_minions_count Number of minions declared in salt
# TYPE saltstack_minions_count gauge
saltstack_minions_count 11
```

Jobs metrics

```
# HELP saltstack_job_status Job status
# TYPE saltstack_job_status gauge
saltstack_job_status{function="state.apply",minion="minion1.local.domain"} 1 1675779556000
saltstack_job_status{function="state.apply",minion="minion2.local.domain"} 1 1675776156000
saltstack_job_status{function="state.apply",minion="minion3.local.domain"} 1 1675764177000
saltstack_job_status{function="state.apply",minion="minion4.local.domain"} 1 1675784608000
saltstack_job_status{function="state.apply",minion="minion5.local.domain"} 1 1675784718000
saltstack_job_status{function="state.apply",minion="minion6.local.domain"} 1 1675783307000
saltstack_job_status{function="state.apply",minion="minion7.local.domain"} 1 1675782089000
saltstack_job_status{function="state.apply",minion="minion8.local.domain"} 1 1675789146000
saltstack_job_status{function="state.apply",minion="minion9.local.domain"} 1 1675780960000
```

## Preparing saltstack API

Before installing the exporter, be sure to install and configure the `salt-api`.

If it's not already in place in your `saltstack` environment, please install the `rest_cherrypy` api.
You can find the installation method [here](https://docs.saltproject.io/en/latest/ref/netapi/all/salt.netapi.rest_cherrypy.html#a-rest-api-for-salt) on the saltstack website.

### create at least one external auth user or group

Edit your `master` configuration file or add a `master.d/external_auth.conf` with the following parameter:

```
external_auth:
  pam:
    ludovic:
      - .*
      - '@runner'
      - '@wheel'
      - '@jobs'
```

This will authenticate the user named "ludovic" via PAM, giving some authorization.
You can try different settings, auth method and authorization, see the [documentation](https://docs.saltproject.io/en/latest/topics/eauth/index.html#acl-eauth)

### rest_cherrypy configuration

You may configure the `salt-api` like this :

```
rest_cherrypy:
  port: 3333
  host: 0.0.0.0
  disable_ssl: true
```

You can set the `disable_ssl` paramater to `false` and use self-signed certificate.

```
salt-call --local tls.create_self_signed_cert
```

And then change the rest_cherrypy configuration like this:

```
rest_cherrypy:
  port: 3333
  host: 0.0.0.0
  disable_ssl: false
  ssl_crt: /etc/pki/tls/certs/localhost.crt
  ssl_key: /etc/pki/tls/certs/localhost.key
```

### testing API

You may have to start or restart the `salt-api` service if it's not already done.

```
systemctl restart salt-api.service
```

Now that the `salt-api` is configured and started we can test it.

#### login

login:

```
curl -sSk http://localhost:3333/login -d username=ludovic -d password="${PASSWORD}" -d eauth=pam
```

return a JSON:

```
{"return": [{"token": "a66e35c1cc59a45aa93f82576f996b6ffbb0d240", "expire": 1674859425.5400722, "start": 1674816225.5400717, "user": "ludovic", "eauth": "pam", "perms": [".*", "@jobs", "@runner", "@wheel"]}]}
```

## Usage

```
usage: prometheus-saltstack-exporter [<flags>]

Flags:
  -h, --help     Show context-sensitive help (also try --help-long and --help-man).
      --config.file="config.yml"
                 Exporter configuration file.
      --web.listen-address=":9142"
                 Address to listen on for telemetry
      --web.telemetry-path="/metrics"
                 Path under which to expose metrics
      --debug    Active debug in log
      --version  Show application version.
```

## Installing

### Debian family

A debian linux package is available for x86_64 platform.

```
dpkg -i prometheus-saltstack-explorer-semver-gitver.deb
```

In the package there are :

* A binary installed in /usr/bin
* A directory /etc/prometheus-saltstack-exporter
* A default `config.yaml` file the configuration directory
* A service file for systemd


### RedHat family

A redhat linux package is available for x86_64 platform.

```
rpm -hiv prometheus-saltstack-explorer-semver-gitver.rpm
```

In the package there are :

* A binary installed in /usr/bin
* A directory /etc/prometheus-saltstack-exporter
* A default `config.yaml` file the configuration directory
* A service file for systemd

### Others

```
go install github.com/Whyrl35/prometheus-saltstack-exporter
```

or download the latest release from [github-releases](https://github.com/Whyrl35/prometheus-saltstack-exporter/releases)

# Development

If you want to contribute, please read the [CONTRIBUTING.md](CONTTRIBUTING.md), then:

1. Use Golang version `>= 1.19`
2. Fork [this repository](https://github.com/Whyrl35/prometheus-saltstack-exporter)
3. Create a feature branch
4. Run test suite with `$ make test` command and be sure it passes
5. Install and user pre-commit `$ pre-commit install`
    * it will need [golangci-lint](https://golangci-lint.run/usage/install/#local-installation)
    * it will need [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)
    * it will need [go-critic](https://github.com/go-critic/go-critic#installation)
    * and it also use [commitlint](https://commitlint.js.org/#/)
7. Edit the code
8. Commit your change
9. Rebase against `master` branch
10. Create a Pull Request

For your commit message, you may follow the concept describe [here](https://mokkapps.de/blog/how-to-automatically-generate-a-helpful-changelog-from-your-git-commit-messages/)

## Install go environment

If you don't already have a Go environment, please follow this [documentation](https://go.dev/doc/install).

## Makefile

A `Makefile` is here to make your life easier to try and build the package.

## Testing

TBD
