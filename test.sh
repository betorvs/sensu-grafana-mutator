#!/usr/bin/env bash

cat $1 | ./sensu-grafana-mutator -g https://grafana.example.com/?orgId=1 -e -k -a \
    -d '[
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
]'