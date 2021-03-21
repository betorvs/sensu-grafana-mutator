
[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/betorvs/sensu-grafana-mutator)
![Go Test](https://github.com/betorvs/sensu-grafana-mutator/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/betorvs/sensu-grafana-mutator/workflows/goreleaser/badge.svg)

# sensu-grafana-mutator

## Table of Contents
- [Overview](#overview)
- [Usage](#usage)
- [Configuration](#configuration)
  - [Requirements](#requirements)
  - [Sensu Kubernetes Events](#sensu-kubernetes-events)
  - [Sensu Alertmanager Events](#sensu-alertmanager-events)
  - [Grafana Dashboard Suggested](#grafana-dashboard-suggested)
    - [Labels and Match Labels](#labels-and-match-labels)
  - [Asset registration](#asset-registration)
  - [Mutator definition](#mutator-definition)
    - [Full Example](#full-example)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

The sensu-grafana-mutator is a [Sensu Mutator][1] created to parse event labels and generate one or more event.check.annotations ending in `_url` with a time range to make sysadmin's life easier to start his troubleshooting. Very basic usage is to parse label and match a Grafana Dashboard with these labels.
It contains 2 builtin integrations: 
- with [sensu-kubernetes-events][4] it can generate one annotation called `grafana_loki_url` parsing kubernetes events labels;
- with [sensu-alertmanager-events][6] it uses `grafana_loki_url` annotation parsing Alert Manager labels. 

## Usage

```bash

Sensu grafana mutator add Grafana Dashboards or Grafana Explore Links in event annotations

Usage:
  sensu-grafana-mutator [flags]
  sensu-grafana-mutator [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -a, --alertmanager-events-integration              Grafana Mutator parser for sensu-alertmanager-events plugin
  -A, --alertmanager-integration-label string        Label used to identify sensu-alertmanager-events plugin events (default "sensu-alertmanager-events")
      --always-return-event                          Grafana Mutator will always return an event, even if it has error. All errors will be reported in event.annotations[sensu-grafana-mutator/error]
      --default-integrations-label-node string       Default node label from Kubernetes Events and Alert Manager integration. (default "node")
      --default-loki-label-hostname string           Default hostname label for Grafana Loki Stream. {hostname=value} (default "hostname")
      --default-loki-label-namespace string          Default namespace label for Grafana Loki Stream. {namespace=value} (default "namespace")
      --extra-loki-labels string                     Extra labels for Grafana Loki Stream. (default "cluster,pod")
  -d, --grafana-dashboard-suggested string           Suggested Dashboard based on Labels and add it in Grafana URL as &var-label[key]=label[value] (only json format). e. [{"grafana_annotation":"kubernetes_namespace","dashboard_url":"https://grafana.example.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos","labels":["namespace"]}]
  -e, --grafana-explore-link-enabled                 Enable Grafana Loki Explore Links
  -D, --grafana-loki-datasource string               An Grafana Loki Datasource name. e. -d loki  (default "loki")
  -r, --grafana-mutator-time-range int               Time range in seconds to create grafana URLs. It will use FromDate = 'event.timestamp - time-range' and ToDate = 'event.timestamp + time-range' (default 300)
  -g, --grafana-url string                           An grafana complete URL. e. https://grafana.com/?orgId=1 
  -h, --help                                         help for sensu-grafana-mutator
  -k, --kubernetes-events-integration                Grafana Mutator parser for sensu-kubernetes-events plugin
      --kubernetes-events-integration-label string   Label used to identify sensu-kubernetes-events plugin events (default "sensu-kubernetes-events")
  -P, --kubernetes-events-pipeline string            Grafana Loki pipeline to match. e. {app=eventrouter} |= io.kubernetes.event.id (default "io.kubernetes.event.id")
  -L, --kubernetes-events-stream-label string        Grafana Loki stream label. e. {app=eventrouter} (default "app")
  -N, --kubernetes-events-stream-namespace string    Grafana Loki stream namespace. e. {app=eventrouter,namespace=io.kubernetes.event.namespace} (default "io.kubernetes.event.namespace")
  -S, --kubernetes-events-stream-selector string     Grafana Loki stream selector. e. {app=eventrouter} (default "eventrouter")
  -s, --sensu-label-selector string                  Sensu Label Selector to create Grafana Explore URL using loki as Datasource. {namespace=kubernetes_namespace.value} (default "kubernetes_namespace")

Use "sensu-grafana-mutator [command] --help" for more information about a command.


```

## Configuration

Basic usage sensu-grafana-mutator should be:

```sh
cat event.json | ./sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -e
```

Output annotation: `event.check.annotations["grafana_loki_url"]`.

To change sensu label selector, use:

```sh
cat event.json | ./sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -e -s namespace
```

### Requirements

You should have [Grafana][8] installed and configured. If you want to use `--grafana-explore-link-enabled` you should have a [Grafana Loki][5] installed and receiving logs. 

### sensu-kubernetes-events

- Import events from kubernetes using [eventrouter][3] to [Grafana Loki][5]
- Import kubernetes events in Sensu using [sensu-kubernetes-events plugin][4]

Using Grafana Explore tab and Grafana Loki as datasource: 
- Search as example: `{app="eventrouter",namespace="default"}|= "nginx-deployment-78dc4549b8-kkxnf.164c27e81b96bdc8"` where "nginx-deployment-78dc4549b8-kkxnf.164c27e81b96bdc8" cames from the value from sensu event.label["io.kubernetes.event.id"].


Then sensu-grafana-mutator should be:

```
cat event.json | ./sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -e -k
```

Output annotation: `event.check.annotations["grafana_loki_url"]`.

### sensu-alertmanager-events

It will try to find the label in event.check.Label with name `sensu-alertmanager-events` and value `owner` then it will create a grafana loki URL using only namespace in stream. Example: `{namespace="Value"}`. Only change `--alertmanager-integration-label` if the [sensu-alertmanager-events][6] plugin changed it.

Then sensu-grafana-mutator should be:

```
cat event.json | ./sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -e -a
```

Output annotation: `event.check.annotations["grafana_loki_url"]`.

### grafana-dashboard-suggested

You can include multiples grafana_annotations inside this flag. But we don't have a benchmark about it. Then keep it simple and it will work as expected. We used one example dashboard from [kubernetes-mixin][7] called kubernetes-compute-resources-namespace-pods. 

```json
[
  {
    "grafana_annotation": "kubernetes_namespace",
    "dashboard_url": "https://grafana.example.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos",
    "labels": [
      "namespace",
      "cluster"
    ]
  }
]
```

But in Sensu yaml configuration should be in one line with scapes:

```bash
cat event.json | ./sensu-grafana-mutator -d "[{\"grafana_annotation\":\"kubernetes_namespace\",\"dashboard_url\":\"https://grafana.example.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos\",\"labels\":[\"namespace\",\"cluster\"]}]"

```

Output annotation: `event.check.annotations["grafana_kubernetes_namespace_url"]`.

#### Labels and Match Labels

In definition for `--grafana-dashboard-suggested` we should use one json (one line with escapes):

```json
[
  {
    "grafana_annotation": "kubernetes_namespace",
    "dashboard_url": "https://grafana.example.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos",
    "labels": [
      "namespace",
      "cluster"
    ]
  },
  {
    "grafana_annotation": "kubelet",
    "dashboard_url": "https://grafana.example.com/d/3138fa155d5915769fbded898ac09fd9/kubernetes-kubelet?orgId=1&var-datasource=thanos",
    "labels": [
      "cluster"
    ],
    "match_labels": {
        "alertname": "KubeletPlegDurationHigh"
    }
  },
  {
    "grafana_annotation": "controller",
    "dashboard_url": "https://grafana.example.com/d/72e0e05bef5099e5f049b05fdc429ed4/kubernetes-controller-manager?orgId=1",
    "match_labels": {
        "alertname": "KubeAPILatencyHigh",
        "component": "apiserver"
    }
  }
]
```

Only to explain it, if:
  - match labels "alertname=KubeletPlegDurationHigh" add: `"grafana_kubelet_url": "https://grafana.example.com/d/3138fa155d5915769fbded898ac09fd9/kubernetes-kubelet?orgId=1&var-datasource=thanos&from=1607077959000&to=1607078559000&var-cluster=k8s-b.dev.ppro.com"`
  - match labels "alertname=KubeAPILatencyHigh" and "component=apiserver" add: `"grafana_controller_url": "https://grafana.example.com/d/72e0e05bef5099e5f049b05fdc429ed4/kubernetes-controller-manager?orgId=1&from=1607412032000&to=1607412332000"`
  - only find these labels "namespace" and "cluster" add: `"grafana_controller_url": "https://grafana.example.com/d/72e0e05bef5099e5f049b05fdc429ed4/kubernetes-controller-manager?orgId=1&from=1607412032000&to=1607412332000"`

### Asset registration

[Sensu Assets][2] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add betorvs/sensu-grafana-mutator
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/betorvs/sensu-grafana-mutator].

### Mutator definition

Basic usage will parse sensu label `kubernetes_namespace` as `namespace` in Grafana Loki Explore URL.

```yml
---
type: Mutator
api_version: core/v2
metadata:
  name: sensu-grafana-mutator
  namespace: default
spec:
  command: sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -e
  runtime_assets:
  - betorvs/sensu-grafana-mutator
```

#### Full example

Will parse sensu events with label `kubernetes_namespace` and events from these plugins [sensu-kubernetes-events][4] and [sensu-alertmanager-events][6] and if found labels `namespace` and `cluster` will add a dashboard [kubernetes-compute-resources-namespace-pods][7] link and if found labels `node` and `cluster` will add dashboard [kubernetes-compute-resources-node-pods][7] link.

```yml
---
type: Mutator
api_version: core/v2
metadata:
  name: sensu-grafana-mutator
  namespace: default
spec:
  command: >-
    sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -e -k -a
    -d "[{\"grafana_annotation\":\"kubernetes_namespace\",\"dashboard_url\":\"https://grafana.example.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos\",\"labels\":[\"namespace\",\"cluster\"]},{\"grafana_annotation\":\"kubernetes_nodes\",\"dashboard_url\":\"https://grafana.example.com/d/200ac8fdbfbb74b39aff88118e4d1c2c/kubernetes-compute-resources-node-pods?orgId=1&var-datasource=thanos\",\"labels\":[\"node\",\"cluster\"]}]"
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

This mutator was design to work with handlers that can parse these grafana annotations as link. Example [sensu-opsgenie-handler][9] and [sensu-hangouts-chat-handler][10].

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
[9]: https://github.com/betorvs/sensu-opsgenie-handler
[10]: https://github.com/betorvs/sensu-hangouts-chat-handler