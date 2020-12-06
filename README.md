
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

The sensu-grafana-mutator is a [Sensu Mutator][1] created to parse event labels and generate one or more event.check.annotations ending in `_url` with a time range to make sysadmin's life easier to start his troubleshooting. Very basic usage is to parse label and match a Grafana Dashboard with these labels.
It contains 2 builtin integrations: 
- with [sensu-kubernetes-events][4] it can generate one annotation called `grafana_loki_url` parsing kubernetes events labels;
- with [sensu-alertmanager-events][6] it uses `grafana_loki_url` annotation parsing Alert Manager labels. 

## Usage examples

```bash

Sensu grafana mutator add Grafana Dashboards or Grafana Explore Links in event annotations

Usage:
  sensu-grafana-mutator [flags]
  sensu-grafana-mutator [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -a, --alertmanager-events-integration             Grafana Mutator parser for sensu-alertmanager-events plugin
  -A, --alertmanager-integration-label string       Label used to identify sensu-alertmanager-events plugin events (default "sensu-alertmanager-events")
  -d, --grafana-dashboard-suggested string          Suggested Dashboard based on Labels. e. [{"grafana_annotation":"kubernetes_namespace","dashboard_url":"https://grafana.example.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos","labels":["namespace"]}]
  -e, --grafana-explore-link-enabled                Enable Grafana Loki Explore Links
  -D, --grafana-loki-datasource string              An Grafana Loki Datasource name. e. -d loki  (default "loki")
  -r, --grafana-mutator-time-range int              Time range in seconds to create grafana URLs (default 300)
  -g, --grafana-url string                          An grafana complete URL. e. https://grafana.com/?orgId=1 
  -h, --help                                        help for sensu-grafana-mutator
  -k, --kubernetes-events-integration               Grafana Mutator parser for sensu-kubernetes-events plugin
  -P, --kubernetes-events-pipeline string           Grafana Loki pipeline to match. e. {app=eventrouter} |= io.kubernetes.event.id (default "io.kubernetes.event.id")
  -L, --kubernetes-events-stream-label string       Grafana Loki stream label. e. {app=eventrouter} (default "app")
  -N, --kubernetes-events-stream-namespace string   Grafana Loki stream namespace. e. {app=eventrouter,namespace=io.kubernetes.event.namespace} (default "io.kubernetes.event.namespace")
  -S, --kubernetes-events-stream-selector string    Grafana Loki stream selector. e. {app=eventrouter} (default "eventrouter")
  -s, --sensu-label-selector string                 Sensu Label Selector to create Grafana Explore URL using loki as Datasource. {namespace=kubernetes_namespace.value} (default "kubernetes_namespace")

Use "sensu-grafana-mutator [command] --help" for more information about a command.


```

## Configuration

### Requirements

You should have [Grafana][8] installed and configured. If you want to use `--grafana-explore-link-enabled` you should have a [Grafana Loki][5] installed and receiving logs. 

#### sensu-kubernetes-events

- Import events from kubernetes using [eventrouter][3] to [Grafana Loki][5]
- Import kubernetes events in Sensu using [sensu-kubernetes-events plugin][4]

Using Grafana Explore tab and Grafana Loki as datasource: 
- Search as example: `{app="eventrouter",namespace="default"}|= "nginx-deployment-78dc4549b8-kkxnf.164c27e81b96bdc8"` where "nginx-deployment-78dc4549b8-kkxnf.164c27e81b96bdc8" cames from the value from sensu event.label["io.kubernetes.event.id"].


Then sensu-grafana-mutator should be:

```
./sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -E
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
[8]: https://grafana.com/docs/grafana/latest/