# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

## [0.0.2] - 

### Added
- Add `--always-return-event` flag to return sensu event with a annotation with this error. To avoid missing any event you should use this flag.
- Add new flags `--extra-loki-labels` and `--kubernetes-events-integration-label` and `--default-integrations-label-node`

### Changed
- change `--grafana-dashboard-suggested` to add match_labels in json
- change `--grafana-mutator-time-range` increate time range
- change grafana loki url generator
- change return event even if orgId is missing from grafana URL in grafana loki url generator


## [0.0.1] - 2020-12-06

### Added
- Initial release: Add Grafana Dashboards URLs, Add Grafana Explore Loki URLs, kubernetes-events plugin integration, alertmanager plugins integration
