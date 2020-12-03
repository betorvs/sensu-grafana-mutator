
[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/betorvs/sensu-grafana-mutator)
![Go Test](https://github.com/betorvs/sensu-grafana-mutator/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/betorvs/sensu-grafana-mutator/workflows/goreleaser/badge.svg)

# sensu-grafana-mutator

## Table of Contents
- [Overview](#overview)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Mutator definition](#mutator-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

The sensu-grafana-mutator is a [Sensu Mutator][1] that parse a label as grafana loki explore url and add to event as grafana_loki_url annotation. 

## Usage examples

```bash

Sensu grafana mutator add grafana_*_url annotations

Usage:
  sensu-grafana-mutator [flags]
  sensu-grafana-mutator [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -A, --alertmanager-integration-label string           Allow integration from sensu-alertmanager-events plugin (default "sensu-alertmanager-events")
  -D, --grafana-dashboard-suggested string              Suggested Dashboard based on Labels. e. [{"grafana_annotation":"kubernetes_namespace","dashboard_url":"https://grafana.example.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos","labels":["namespace"]}]
  -d, --grafana-loki-datasource string                  An Grafana Loki Datasource name. e. -d loki  (default "loki")
  -p, --grafana-loki-explorer-pipeline string           From Sensu Events, choose one label to be parse here. e. {app=eventrouter} |= k8s_id then use -p k8s_id
  -r, --grafana-loki-explorer-range int                 Time range in seconds to create grafana explorer URL (default 300)
  -l, --grafana-loki-explorer-stream-label string       From Grafana Loki streams use label. e. {app=eventrouter} then '-l app'  (default "app")
  -n, --grafana-loki-explorer-stream-namespace string   From Grafana Loki streams use namespace. e. {namespace=ValueFromEvent} then '-n NamespaceLabelName' 
  -s, --grafana-loki-explorer-stream-selector string    From Grafana Loki streams use label. e. {app=eventrouter} then '-s eventrouter'  (default "eventrouter")
  -g, --grafana-url string                              An grafana complete URL. e. https://grafana.com/?orgId=1 
  -h, --help                                            help for sensu-grafana-mutator

Use "sensu-grafana-mutator [command] --help" for more information about a command.

```

## Configuration

### Requirements

- Import events from kubernetes using [eventrouter][3] to [Grafana Loki][5]
- Import kubernetes events in Sensu using [sensu-kubernetes-events plugin][4]

Using Grafana Explore tab and Grafana Loki as datasource: 
- Search as example: `{app="eventrouter"}|= "164c27e81b96bdc8"` where "164c27e81b96bdc8" cames from the value from sensu event.label["io.kubernetes.event.id"].


Then sensu-grafana-mutator should be:

```
./sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -p io.kubernetes.event.id
```

#### sensu-alertmanager-events


It will try to find the label in event.check.Label with name `sensu-alertmanager-events` and value `owner` then it will create a grafana loki URL using only namespace in stream. Example: `{namespace="Value"}`. Only change `--alertmanager-integration-label` if the [sensu-alertmanager-events][6] plugin changed it.

#### grafana-dashboard-suggested

You can include multiples grafana_annotations inside this flag. But we don't have a benchmark about it. Then keep it simple and it will work as expected. We used one example dashboard from [kubernete-mixin][7] called kubernetes-compute-resources-namespace-pods. 

```json
[
  {
    "grafana_annotation": "kubernetes_namespace",
    "dashboard_url": "https://grafana-beta.k8s.infra.ppro.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos",
    "labels": [
      "namespace",
      "cluster"
    ]
  }
]
```

### Asset registration

[Sensu Assets][2] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add betorvs/sensu-grafana-mutator
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/betorvs/sensu-grafana-mutator].

### Mutator definition

```yml
---
type: Mutator
api_version: core/v2
metadata:
  name: sensu-grafana-mutator
  namespace: default
spec:
  command: sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -p io.kubernetes.event.id
  runtime_assets:
  - betorvs/sensu-grafana-mutator
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-grafana-mutator repository:

```
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://docs.sensu.io/sensu-go/latest/reference/mutators/
[2]: https://docs.sensu.io/sensu-go/latest/reference/assets/
[3]: https://github.com/heptiolabs/eventrouter
[4]: https://github.com/betorvs/sensu-kubernetes-events
[5]: https://grafana.com/docs/loki/latest
[6]: https://github.com/betorvs/sensu-alertmanager-events
[7]: https://github.com/kubernetes-monitoring/kubernetes-mixin