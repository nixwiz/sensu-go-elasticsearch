# Sensu Go Elasticsearch metric handler plugin
[![Bonsai Asset Badge](https://img.shields.io/badge/Sensu%20Go%20Elasticsearch-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/nixwiz/sensu-go-elasticsearch)[![Go Test](https://github.com/nixwiz/sensu-go-elasticsearch/workflows/Go%20Test/badge.svg)](https://github.com/nixwiz/sensu-go-elasticsearch/actions?query=workflow%3A%22Go+Test%22)[![Go Lint](https://github.com/nixwiz/sensu-go-elasticsearch/workflows/Go%20Lint/badge.svg)](https://github.com/nixwiz/sensu-go-elasticsearch/actions?query=workflow%3A%22Go+Lint%22)[![goreleaser](https://github.com/nixwiz/sensu-go-elasticsearch/workflows/goreleaser/badge.svg)](https://github.com/nixwiz/sensu-go-elasticsearch/actions?query=workflow%3A%22goreleaser%22)

## Table of Contents
- [Overview](#overview)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Handler definition](#handler-definition)
  - [Proxy support](#proxy-support)
  - [Annotations](#annotations)
- [Installation from source](#installation-from-source)
- [Contributing](#contributing)

## Overview

sensu-go-elasticsearch is a [Sensu Handler][2] for pushing metric data and full
event bodies into Elasticsearch for visualization in Kibana.

## Usage Examples

Help:

```
The Sensu Go handler for metric and event logging in elasticsearch
Required:  Set the ELASTICSEARCH_URL env var with an appropriate connection url (https://user:pass@hostname:port)

Usage:
  sensu-go-elasticsearch [flags]
  sensu-go-elasticsearch [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -d, --dated_index                 Should the index have the current date postfixed? ie: metric_data-2019-06-27
  -f, --full_event_logging          send the full event body instead of isolating event metrics
  -h, --help                        help for sensu-go-elasticsearch
  -i, --index string                index to use
  -s, --insecure-skip-verify        skip TLS certificate verification (not recommended!)
  -p, --point_name_as_metric_name   use the entire point name as the metric name
  -t, --trusted-ca-file string      TLS CA certificate bundle in PEM format

Use "sensu-go-elasticsearch [command] --help" for more information about a command.
```

## Configuration

### Asset registration

[Sensu Assets][3] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add nixwiz/sensu-go-elasticsearch
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][4]

### Handler definition

#### For metrics only

```yml
---
type: Handler
api_version: core/v2
metadata:
  name: sensu-go-elasticsearch
  namespace: default
spec:
  command: sensu-go-elasticsearch -i sensu --dated_index
  type: pipe
  runtime_assets:
    - nixwiz/sensu-go-elasticsearch
  secrets:
    - name: ELASTICSEARCH_URL
      secret: elasticsearch_urL
  filters:
    - has_metrics
```

#### For full event data

```yml
---
type: Handler
api_version: core/v2
metadata:
  name: sensu-go-elasticsearch
  namespace: default
spec:
  command: sensu-go-elasticsearch -i sensu --dated_index --full_event_logging
  type: pipe
  runtime_assets:
    - nixwiz/sensu-go-elasticsearch
  secrets:
    - name: ELASTICSEARCH_URL
      secret: elasticsearch_urL
```

### Environment Variables

The handler requires the ELASTICSEARCH_URL environment variable.  Given that the
username/password may be included in this URL, it is suggested to make use of
[secrets management][5] to surface it.  The handler definitions above reference
it as a secret.  Below is an example secret definition that makes use of the
built-in [env secrets provider][6].

```yml
---
type: Secret
api_version: secrets/v1
metadata:
  name: elasticsearch_url
spec:
  provider: env
  id: ELASTICSEARCH_URL
```

#### Proxy support

This handler supports the use of the environment variables HTTP_PROXY,
HTTPS_PROXY, and NO_PROXY (or the lowercase versions thereof). HTTPS_PROXY takes
precedence over HTTP_PROXY for https requests.  The environment values may be
either a complete URL or a "host[:port]", in which case the "http" scheme is assumed.

### Annotations

All arguments for this handler are tunable on a per entity or check basis based on annotations.  The
annotations keyspace for this handler is `sensu.io/plugins/elasticsearch/config`.

#### Examples

To change the index argument for a particular entity, in that entity's agent.yml add the following:

```yml
[...]
annotations:
  sensu.io/plugins/elasticsearch/config: "dev_index"
[...]
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable from this source.

From the local path of the sensu-go-elasticsearch repository:

```
go build
```

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://docs.sensu.io/sensu-go/latest/reference/handlers/
[3]: https://docs.sensu.io/sensu-go/latest/reference/assets/
[4]: https://bonsai.sensu.io/assets/nixwiz/sensu-go-elasticsearch].
[5]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/
[6]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/#use-env-for-secrets-management

