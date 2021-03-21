package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// DashboardSuggested struct
type DashboardSuggested struct {
	GrafanaAnnotation string            `json:"grafana_annotation"`
	DashboardURL      string            `json:"dashboard_url"`
	Labels            []string          `json:"labels"`
	MatchLabels       map[string]string `json:"match_labels"`
}

// Config represents the mutator plugin config.
type Config struct {
	sensu.PluginConfig
	GrafanaURL                      string
	GrafanaDashboardSuggested       string
	GrafanaExploreLinkEnabled       bool
	GrafanaLokiDatasource           string
	SensuLabelSelector              string
	KubernetesIntegrationLabel      string
	KubernetesEventsIntegration     bool
	KubernetesEventsStreamLabel     string
	KubernetesEventsStreamSelector  string
	KubernetesEventsPipeline        string
	KubernetesEventsStreamNamespace string
	AlertmanagerEventsIntegration   bool
	AlertmanagerIntegrationLabel    string
	DefaultLokiLabelNamespace       string
	DefaultLokiLabelHostname        string
	DefaultIntegrationsLabelNode    string
	ExtraLokiLabels                 string
	AlwaysReturnEvent               bool
	GrafanaMutatorTimeRange         int
	TimeRange                       int64
}

var (
	mutatorConfig = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-grafana-mutator",
			Short:    "Sensu grafana mutator add Grafana Dashboards or Grafana Explore Links in event annotations",
			Keyspace: "sensu.io/plugins/sensu-grafana-mutator/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "grafana-url",
			Env:       "GRAFANA_URL",
			Argument:  "grafana-url",
			Shorthand: "g",
			Default:   "",
			Usage:     "An grafana complete URL. e. https://grafana.com/?orgId=1 ",
			Value:     &mutatorConfig.GrafanaURL,
		},
		{
			Path:      "grafana-dashboard-suggested",
			Env:       "",
			Argument:  "grafana-dashboard-suggested",
			Shorthand: "d",
			Default:   "",
			Usage:     "Suggested Dashboard based on Labels and add it in Grafana URL as &var-label[key]=label[value] (only json format). e. [{\"grafana_annotation\":\"kubernetes_namespace\",\"dashboard_url\":\"https://grafana.example.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos\",\"labels\":[\"namespace\"]}]",
			Value:     &mutatorConfig.GrafanaDashboardSuggested,
		},
		{
			Path:      "grafana-explore-link-enabled",
			Env:       "",
			Argument:  "grafana-explore-link-enabled",
			Shorthand: "e",
			Default:   false,
			Usage:     "Enable Grafana Loki Explore Links",
			Value:     &mutatorConfig.GrafanaExploreLinkEnabled,
		},
		{
			Path:      "kubernetes-events-integration",
			Env:       "",
			Argument:  "kubernetes-events-integration",
			Shorthand: "k",
			Default:   false,
			Usage:     "Grafana Mutator parser for sensu-kubernetes-events plugin",
			Value:     &mutatorConfig.KubernetesEventsIntegration,
		},
		{
			Path:      "alertmanager-events-integration",
			Env:       "",
			Argument:  "alertmanager-events-integration",
			Shorthand: "a",
			Default:   false,
			Usage:     "Grafana Mutator parser for sensu-alertmanager-events plugin",
			Value:     &mutatorConfig.AlertmanagerEventsIntegration,
		},
		{
			Path:      "always-return-event",
			Env:       "",
			Argument:  "always-return-event",
			Shorthand: "",
			Default:   false,
			Usage:     "Grafana Mutator will always return an event, even if it has error. All errors will be reported in event.annotations[sensu-grafana-mutator/error]",
			Value:     &mutatorConfig.AlwaysReturnEvent,
		},
		{
			Path:      "grafana-mutator-time-range",
			Env:       "",
			Argument:  "grafana-mutator-time-range",
			Shorthand: "r",
			Default:   300,
			Usage:     "Time range in seconds to create grafana URLs. It will use FromDate = 'event.timestamp - time-range' and ToDate = 'event.timestamp + time-range'",
			Value:     &mutatorConfig.GrafanaMutatorTimeRange,
		},
		{
			Path:      "grafana-loki-datasource",
			Env:       "GRAFANA_LOKI_DATASOURCE",
			Argument:  "grafana-loki-datasource",
			Shorthand: "D",
			Default:   "loki",
			Usage:     "An Grafana Loki Datasource name. e. -d loki ",
			Value:     &mutatorConfig.GrafanaLokiDatasource,
		},
		{
			Path:      "sensu-label-selector",
			Env:       "SENSU_LABEL_SELECTOR",
			Argument:  "sensu-label-selector",
			Shorthand: "s",
			Default:   "kubernetes_namespace",
			Usage:     "Sensu Label Selector to create Grafana Explore URL using loki as Datasource. {namespace=kubernetes_namespace.value}",
			Value:     &mutatorConfig.SensuLabelSelector,
		},
		{
			Path:      "alertmanager-integration-label",
			Env:       "",
			Argument:  "alertmanager-integration-label",
			Shorthand: "A",
			Default:   "sensu-alertmanager-events",
			Usage:     "Label used to identify sensu-alertmanager-events plugin events",
			Value:     &mutatorConfig.AlertmanagerIntegrationLabel,
		},
		{
			Path:      "kubernetes-events-integration-label",
			Env:       "",
			Argument:  "kubernetes-events-integration-label",
			Shorthand: "",
			Default:   "sensu-kubernetes-events",
			Usage:     "Label used to identify sensu-kubernetes-events plugin events",
			Value:     &mutatorConfig.KubernetesIntegrationLabel,
		},
		{
			Path:      "kubernetes-events-stream-label",
			Env:       "KUBERNETES_EVENTS_STREAM_LABEL",
			Argument:  "kubernetes-events-stream-label",
			Shorthand: "L",
			Default:   "app",
			Usage:     "Grafana Loki stream label. e. {app=eventrouter}",
			Value:     &mutatorConfig.KubernetesEventsStreamLabel,
		},
		{
			Path:      "kubernetes-events-stream-selector",
			Env:       "KUBERNETES_EVENTS_STREAM_SELECTOR",
			Argument:  "kubernetes-events-stream-selector",
			Shorthand: "S",
			Default:   "eventrouter",
			Usage:     "Grafana Loki stream selector. e. {app=eventrouter}",
			Value:     &mutatorConfig.KubernetesEventsStreamSelector,
		},
		{
			Path:      "kubernetes-events-pipeline",
			Env:       "KUBERNETES_EVENTS_PIPELINE",
			Argument:  "kubernetes-events-pipeline",
			Shorthand: "P",
			Default:   "io.kubernetes.event.id",
			Usage:     "Grafana Loki pipeline to match. e. {app=eventrouter} |= io.kubernetes.event.id",
			Value:     &mutatorConfig.KubernetesEventsPipeline,
		},
		{
			Path:      "kubernetes-events-stream-namespace",
			Env:       "KUBERNETES_EVENTS_STREAM_NAMESPACE",
			Argument:  "kubernetes-events-stream-namespace",
			Shorthand: "N",
			Default:   "io.kubernetes.event.namespace",
			Usage:     "Grafana Loki stream namespace. e. {app=eventrouter,namespace=io.kubernetes.event.namespace}",
			Value:     &mutatorConfig.KubernetesEventsStreamNamespace,
		},
		{
			Path:      "default-loki-label-namespace",
			Env:       "DEFAULT_LOKI_LABEL_NAMESPACE",
			Argument:  "default-loki-label-namespace",
			Shorthand: "",
			Default:   "namespace",
			Usage:     "Default namespace label for Grafana Loki Stream. {namespace=value}",
			Value:     &mutatorConfig.DefaultLokiLabelNamespace,
		},
		{
			Path:      "default-loki-label-hostname",
			Env:       "DEFAULT_LOKI_LABEL_HOSTNAME",
			Argument:  "default-loki-label-hostname",
			Shorthand: "",
			Default:   "hostname",
			Usage:     "Default hostname label for Grafana Loki Stream. {hostname=value}",
			Value:     &mutatorConfig.DefaultLokiLabelHostname,
		},
		{
			Path:      "default-integrations-label-node",
			Env:       "DEFAULT_INTEGRATIONS_LABEL_NODE",
			Argument:  "default-integrations-label-node",
			Shorthand: "",
			Default:   "node",
			Usage:     "Default node label from Kubernetes Events and Alert Manager integration.",
			Value:     &mutatorConfig.DefaultIntegrationsLabelNode,
		},
		{
			Path:      "extra-loki-labels",
			Env:       "EXTRA_LOKI_LABELS",
			Argument:  "extra-loki-labels",
			Shorthand: "",
			Default:   "cluster,pod",
			Usage:     "Extra labels for Grafana Loki Stream.",
			Value:     &mutatorConfig.ExtraLokiLabels,
		},
	}
)

func main() {
	mutator := sensu.NewGoMutator(&mutatorConfig.PluginConfig, options, checkArgs, executeMutator)
	mutator.Execute()
}

func checkArgs(_ *types.Event) error {
	if mutatorConfig.GrafanaDashboardSuggested == "" && !mutatorConfig.GrafanaExploreLinkEnabled {
		return fmt.Errorf("please choose one of these two flags --grafana-dashboard-suggested or --grafana-explore-link-enabled")
	}
	if mutatorConfig.GrafanaExploreLinkEnabled && mutatorConfig.GrafanaURL == "" {
		return fmt.Errorf("using --grafana-explore-link-enabled then --grafana-url or GRAFANA_URL environment variable is required")
	}
	mutatorConfig.TimeRange = int64(mutatorConfig.GrafanaMutatorTimeRange * 1000)
	return nil
}

func executeMutator(event *types.Event) (*types.Event, error) {
	// log.Println("executing mutator with --grafana-url", mutatorConfig.GrafanaURL)
	annotations := make(map[string]string)
	fromDate := event.Timestamp*1000 - mutatorConfig.TimeRange
	toDate := event.Timestamp*1000 + mutatorConfig.TimeRange
	errorAnnotationName := fmt.Sprintf("%s/error", mutatorConfig.Name)
	// if check.annotations is empty, make it
	if event.Check.Annotations == nil {
		event.Check.Annotations = make(map[string]string)
	}
	// to create grafana_loki_url annotation
	if mutatorConfig.GrafanaExploreLinkEnabled {
		labels := labelsToSearch()
		extractedLabels, othersIntegrationsFound := extractLokiLabels(event, labels)
		// using sensu-kubernetes-events plugin
		if mutatorConfig.KubernetesEventsIntegration && othersIntegrationsFound {
			grafanaURL, err := generateGrafanaURL(extractedLabels, fromDate, toDate)
			if err != nil {
				annotations[errorAnnotationName] = fmt.Sprintf("failed generating grafana URL %v", err)
				event.Check.Annotations = mergeStringMaps(event.Check.Annotations, annotations)
				if mutatorConfig.AlwaysReturnEvent {
					return event, nil
				}
				return event, err
			}
			annotations["grafana_loki_url"] = grafanaURL
		}
		// using sensu-alertmanager-events plugin
		if mutatorConfig.AlertmanagerEventsIntegration && othersIntegrationsFound {
			grafanaURL, err := generateGrafanaURL(extractedLabels, fromDate, toDate)
			if err != nil {
				annotations[errorAnnotationName] = fmt.Sprintf("failed generating grafana URL %v", err)
				event.Check.Annotations = mergeStringMaps(event.Check.Annotations, annotations)
				if mutatorConfig.AlwaysReturnEvent {
					return event, nil
				}
				return event, err
			}
			annotations["grafana_loki_url"] = grafanaURL
		}
		// using sensu label defined in --sensu-label-selector
		if !othersIntegrationsFound {
			grafanaURL, err := generateGrafanaURL(extractedLabels, fromDate, toDate)
			if err != nil {
				annotations[errorAnnotationName] = fmt.Sprintf("failed generating grafana URL %v", err)
				event.Check.Annotations = mergeStringMaps(event.Check.Annotations, annotations)
				if mutatorConfig.AlwaysReturnEvent {
					return event, nil
				}
				return event, err
			}
			annotations["grafana_loki_url"] = grafanaURL
		}

	}
	// add any dashboard configured in --grafana-dashboard-suggested
	if mutatorConfig.GrafanaDashboardSuggested != "" {
		dashboardSuggested := []DashboardSuggested{}
		err := json.Unmarshal([]byte(mutatorConfig.GrafanaDashboardSuggested), &dashboardSuggested)
		if err != nil {
			annotations[errorAnnotationName] = fmt.Sprintf("json config %v", err)
			event.Check.Annotations = mergeStringMaps(event.Check.Annotations, annotations)
			if mutatorConfig.AlwaysReturnEvent {
				return event, nil
			}
			return event, err
		}
		for _, v := range dashboardSuggested {
			output := fmt.Sprintf("grafana_%s_url", strings.ToLower(v.GrafanaAnnotation))
			grafanaURL, err := url.Parse(v.DashboardURL)
			if err != nil {
				annotations[errorAnnotationName] = fmt.Sprintf("failed generating grafana URL %v", err)
				event.Check.Annotations = mergeStringMaps(event.Check.Annotations, annotations)
				if mutatorConfig.AlwaysReturnEvent {
					return event, nil
				}
				return event, err
			}
			if !checkMissingOrgID(grafanaURL.Query()) {
				annotations[errorAnnotationName] = "Missing orgId in grafana URL in --grafana-dashboard-suggested. e. https://grafana.com/?orgId=1"
				event.Check.Annotations = mergeStringMaps(event.Check.Annotations, annotations)
				if mutatorConfig.AlwaysReturnEvent {
					return event, nil
				}
				return event, fmt.Errorf("Missing orgId in grafana URL in --grafana-dashboard-suggested. e. https://grafana.com/?orgId=1")
			}
			timeRange := fmt.Sprintf("&from=%d&to=%d", fromDate, toDate)
			if v.MatchLabels != nil {
				if searchMatchLabels(event, v.MatchLabels) {
					if v.Labels != nil {
						// case match matchLabels and found labels
						finalURI, validFinalURI := generateURIBySlice(event, v.Labels)
						if validFinalURI {
							annotations[output] = fmt.Sprintf("%s%s%s", grafanaURL, timeRange, finalURI)
						}
					} else {
						// only match labels is used, no labels provided
						annotations[output] = fmt.Sprintf("%s%s", grafanaURL, timeRange)
					}
				}

			} else {
				finalURI, validFinalURI := generateURIBySlice(event, v.Labels)
				if validFinalURI {
					annotations[output] = fmt.Sprintf("%s%s%s", grafanaURL, timeRange, finalURI)
				}
			}
		}

	}

	// merge new annotations into event.check.annotation
	event.Check.Annotations = mergeStringMaps(event.Check.Annotations, annotations)

	return event, nil
}

func generateGrafanaURL(l map[string]string, fromDate, toDate int64) (string, error) {
	grafanaURL, err := grafanaExploreURLEncoded(l, mutatorConfig.GrafanaURL, mutatorConfig.GrafanaLokiDatasource, fromDate, toDate)
	if err != nil {
		return "", err
	}
	return grafanaURL, nil
}

func replaceSpecial(s string) string {
	//  [
	value := strings.ReplaceAll(s, "[", "%5B")
	//  ] %5D
	value = strings.ReplaceAll(value, "]", "%5D")
	//  " %22
	value = strings.ReplaceAll(value, "\"", "%22")
	// { %7B
	value = strings.ReplaceAll(value, "{", "%7B")
	// } %7D
	value = strings.ReplaceAll(value, "}", "%7D")
	return value
}

func grafanaExploreURLEncoded(labels map[string]string, grafana, datasource string, fromDate, toDate int64) (string, error) {
	// grafana URL expected: https://grafana.com/?orgId=1
	grafanaURL, err := url.Parse(grafana)
	if err != nil {
		return "", err
	}
	// if grafana URL not contain "?orgId=1" return a error in the end of this func
	var errOrgID error
	if !checkMissingOrgID(grafanaURL.Query()) {
		errOrgID = fmt.Errorf("Missing orgId in grafana URL. e. https://grafana.com/?orgId=1")
	}
	grafanaURL.Path = "explore"
	grafanaExploreURL := fmt.Sprintf("%s&left=", grafanaURL)
	var labelsSearchText, startSearchText, endSearchText string
	var eventID bool
	count := 0
	for key, value := range labels {
		if key != "" && value != "" && key != "eventID" {
			if count == 0 {
				labelsSearchText += fmt.Sprintf("%s=\\\"%s\\\"", key, value)
				count++
			} else {
				labelsSearchText += fmt.Sprintf(",%s=\\\"%s\\\"", key, value)
				count++
			}
		}
		if key == "eventID" && value != "" {
			endSearchText = fmt.Sprintf("|=\\\"%s\\\"", value)
			eventID = true
			count++
		}
	}
	startSearchText = fmt.Sprintf("{%s}", labelsSearchText)
	searchText := url.QueryEscape(startSearchText)
	if eventID {
		searchText = url.QueryEscape(fmt.Sprintf("%s%s", startSearchText, endSearchText))
	}
	grafanaExploreURI := fmt.Sprintf("[\"%d\",\"%d\",\"%s\",{\"expr\":\"%s\"}]", fromDate, toDate, datasource, searchText)
	result := fmt.Sprintf("%s%s", grafanaExploreURL, replaceSpecial(grafanaExploreURI))
	return result, errOrgID
}

func checkMissingOrgID(u url.Values) bool {
	for k, v := range u {
		if k == "orgId" && len(v) != 0 {
			return true
		}
	}
	return false
}

func extractLabels(event *types.Event, label string) (string, bool) {
	labelFound := ""
	if event.Labels != nil {
		for k, v := range event.Labels {
			if k == label {
				labelFound = v
			}
		}
	}
	if event.Entity.Labels != nil {
		for k, v := range event.Entity.Labels {
			if k == label {
				labelFound = v
			}
		}
	}
	if event.Check.Labels != nil {
		for k, v := range event.Check.Labels {
			if k == label {
				labelFound = v
			}
		}
	}
	if labelFound == "" {
		return labelFound, false
	}
	return labelFound, true
}

func labelsToSearch() []string {
	labels := stringToSliceStrings(mutatorConfig.ExtraLokiLabels)
	if mutatorConfig.KubernetesEventsIntegration {
		labels = append(labels, mutatorConfig.KubernetesEventsStreamNamespace)
		labels = append(labels, mutatorConfig.KubernetesEventsPipeline)
	}
	labels = append(labels, mutatorConfig.DefaultLokiLabelNamespace)
	if mutatorConfig.SensuLabelSelector != mutatorConfig.DefaultLokiLabelNamespace {
		labels = append(labels, mutatorConfig.SensuLabelSelector)
	}
	if mutatorConfig.AlertmanagerEventsIntegration {
		labels = append(labels, mutatorConfig.DefaultIntegrationsLabelNode)
	}
	return labels
}

func renameKey(s string) string {
	switch {
	case s == mutatorConfig.KubernetesEventsStreamNamespace:
		return mutatorConfig.DefaultLokiLabelNamespace
	case s == mutatorConfig.KubernetesEventsPipeline:
		return "eventID"
	case s == mutatorConfig.SensuLabelSelector:
		return mutatorConfig.DefaultLokiLabelNamespace
	case s == mutatorConfig.DefaultIntegrationsLabelNode:
		return mutatorConfig.DefaultLokiLabelHostname
	default:
		return s
	}
}

func extractLokiLabels(event *types.Event, labels []string) (map[string]string, bool) {
	labelsFound := make(map[string]string)
	var othersIntegrationsFound bool
	for _, l := range labels {
		// [mutatorConfig.KubernetesIntegrationLabel] == "owner"
		if event.Labels != nil {
			for k, v := range event.Labels {
				if k == l {
					key := renameKey(k)
					labelsFound[key] = v
				}
				if k == mutatorConfig.KubernetesIntegrationLabel && v == "owner" {
					othersIntegrationsFound = true
					labelsFound[mutatorConfig.KubernetesEventsStreamLabel] = mutatorConfig.KubernetesEventsStreamSelector
				}
			}
		}
		if event.Entity.Labels != nil {
			for k, v := range event.Entity.Labels {
				if k == l {
					key := renameKey(k)
					labelsFound[key] = v
				}
			}
		}
		// [mutatorConfig.AlertmanagerIntegrationLabel] == "owner"
		if event.Check.Labels != nil {
			for k, v := range event.Check.Labels {
				if k == l {
					key := renameKey(k)
					value := v
					if k == mutatorConfig.DefaultIntegrationsLabelNode {
						// if doesnt find namespace in labels, use hostname = node
						// in Loki every node is labeled as hostname
						// in alert manager/kubernetes the label is node and it used a FQDN
						// example: ip-10-192-172-1.eu-west-1.compute.internal
						if strings.Contains(value, ".") {
							newapp := strings.Split(value, ".")
							value = newapp[0]
						}
					}
					labelsFound[key] = value
				}
				if k == mutatorConfig.AlertmanagerIntegrationLabel && v == "owner" {
					othersIntegrationsFound = true
				}
			}
		}
	}
	return labelsFound, othersIntegrationsFound
}

func generateURIBySlice(event *types.Event, v []string) (string, bool) {
	count := 0
	finalURI := ""
	for _, s := range v {
		// &var-namespace=test
		value, validFinalURI := extractLabels(event, s)
		if validFinalURI {
			finalURI += fmt.Sprintf("&var-%s=%s", s, value)
			count++
		}
	}
	if len(v) == count {
		return finalURI, true
	}
	return "", false
}

func searchMatchLabels(event *types.Event, labels map[string]string) bool {
	if len(labels) == 0 {
		return false
	}
	count := 0
	for key, value := range labels {
		if event.Labels != nil {
			for k, v := range event.Labels {
				if k == key && v == value {
					count++
				}
			}
		}
		if event.Entity.Labels != nil {
			for k, v := range event.Entity.Labels {
				if k == key && v == value {
					count++
				}
			}
		}
		if event.Check.Labels != nil {
			for k, v := range event.Check.Labels {
				if k == key && v == value {
					count++
				}
			}
		}
		if count == len(labels) {
			return true
		}
	}

	return false
}

func mergeStringMaps(left, right map[string]string) map[string]string {
	for k, v := range right {
		// fmt.Println(left[k])
		if left[k] == "" {
			left[k] = v
		}
	}
	return left
}

func stringToSliceStrings(s string) []string {
	slice := []string{}
	if s != "" {
		if strings.Contains(s, ",") {
			splited := strings.Split(s, ",")
			for _, v := range splited {
				if v != "" {
					slice = append(slice, v)
				}
			}
		} else {
			slice = []string{s}
		}
	}
	return slice
}
