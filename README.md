# salt-exporter

## Description

TBD: describe the exporter

## Metrics

```
list all metrics here
```

## Preparing saltstack API

### rest_cherrypy

### testing API

#### login

login:

```
curl -sSk http://${URL}/login -d username=${USERNAME} -d password="${PASSWORD}" -d eauth=pam
```

return a JSON:

```
{"return": [{"token": "a66e35c1cc59a45aa93f82576f996b6ffbb0d240", "expire": 1674859425.5400722, "start": 1674816225.5400717, "user": "_username_", "eauth": "pam", "perms": [".*", "@jobs", "@runner", "@wheel"]}]}
```

## Usage

## Installing

### Debian family

### RedHat family

### Others

# Development

## Install go environment

## Makefile

## Testing
