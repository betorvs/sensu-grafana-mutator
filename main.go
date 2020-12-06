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
	GrafanaAnnotation string   `json:"grafana_annotation"`
	DashboardURL      string   `json:"dashboard_url"`
	Labels            []string `json:"labels"`
}

// Config represents the mutator plugin config.
type Config struct {
	sensu.PluginConfig
	GrafanaURL                      string
	GrafanaDashboardSuggested       string
	GrafanaExploreLinkEnabled       bool
	GrafanaLokiDatasource           string
	SensuLabelSelector              string
	KubernetesEventsIntegration     bool
	KubernetesEventsStreamLabel     string
	KubernetesEventsStreamSelector  string
	KubernetesEventsPipeline        string
	KubernetesEventsStreamNamespace string
	AlertmanagerEventsIntegration   bool
	AlertmanagerIntegrationLabel    string
	DefaultLokiLabelNamespace       string
	DefaultLokiLabelHostname        string
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
			Usage:     "Suggested Dashboard based on Labels. e. [{\"grafana_annotation\":\"kubernetes_namespace\",\"dashboard_url\":\"https://grafana.example.com/d/85a562078cdf77779eaa1add43ccec1e/kubernetes-compute-resources-namespace-pods?orgId=1&var-datasource=thanos\",\"labels\":[\"namespace\"]}]",
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
			Path:      "grafana-mutator-time-range",
			Env:       "",
			Argument:  "grafana-mutator-time-range",
			Shorthand: "r",
			Default:   300,
			Usage:     "Time range in seconds to create grafana URLs",
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
	}
)

func main() {
	mutator := sensu.NewGoMutator(&mutatorConfig.PluginConfig, options, checkArgs, executeMutator)
	mutator.Execute()
}

func checkArgs(_ *types.Event) error {
	if mutatorConfig.GrafanaURL == "" {
		return fmt.Errorf("--grafana-url or GRAFANA_URL environment variable is required")
	}
	mutatorConfig.TimeRange = int64(mutatorConfig.GrafanaMutatorTimeRange * 1000)
	return nil
}

func executeMutator(event *types.Event) (*types.Event, error) {
	// log.Println("executing mutator with --grafana-url", mutatorConfig.GrafanaURL)
	annotations := make(map[string]string)
	fromDate := event.Timestamp * 1000
	toDate := event.Timestamp*1000 + mutatorConfig.TimeRange
	// to create grafana_loki_url annotation
	if mutatorConfig.GrafanaExploreLinkEnabled {
		sensuLabel, sensuLabelExist := extractLabels(event, mutatorConfig.SensuLabelSelector)
		// using sensu-kubernetes-events plugin
		if mutatorConfig.KubernetesEventsIntegration && !sensuLabelExist {
			ExplorePipeline, ExploreValid := extractLabels(event, mutatorConfig.KubernetesEventsPipeline)
			namespace := ""
			namespaceStream, namespaceValid := extractLabels(event, mutatorConfig.KubernetesEventsStreamNamespace)
			if namespaceValid {
				namespace = namespaceStream
			}
			if ExploreValid {
				label := mutatorConfig.KubernetesEventsStreamLabel
				app := mutatorConfig.KubernetesEventsStreamSelector
				grafanaURL, err := generateGrafanaURL(label, app, ExplorePipeline, namespace, fromDate, toDate)
				if err != nil {
					return event, err
				}
				annotations["grafana_loki_url"] = grafanaURL
			}
		}
		// using sensu-alertmanager-events plugin
		if event.Check.Labels[mutatorConfig.AlertmanagerIntegrationLabel] == "owner" && !sensuLabelExist {
			app, nameValid := extractLabels(event, mutatorConfig.DefaultLokiLabelNamespace)
			if nameValid {
				grafanaURL, err := generateGrafanaURL(mutatorConfig.DefaultLokiLabelNamespace, app, "", "", fromDate, toDate)
				if err != nil {
					return event, err
				}
				annotations["grafana_loki_url"] = grafanaURL
			}
			// if doesnt find namespace in labels, use hostname = node
			// in Loki every node is labeled as hostname
			// in alert manager/kubernetes the label is node and it used a FQDN
			// ip-10-192-172-1.eu-west-1.compute.internal
			if !nameValid {
				label := mutatorConfig.DefaultLokiLabelHostname
				app, nameValid := extractLabels(event, "node")
				if nameValid {
					// parse FQDN and use short hostname
					if strings.Contains(app, ".") {
						newapp := strings.Split(app, ".")
						app = newapp[0]
					}
					grafanaURL, err := generateGrafanaURL(label, app, "", "", fromDate, toDate)
					if err != nil {
						return event, err
					}
					annotations["grafana_loki_url"] = grafanaURL
				}
			}
		}
		// using sensu label defined in --sensu-label-selecto
		if sensuLabelExist {
			label := mutatorConfig.DefaultLokiLabelNamespace
			if mutatorConfig.SensuLabelSelector != "kubernetes_namespace" {
				label = mutatorConfig.SensuLabelSelector
			}
			grafanaURL, err := generateGrafanaURL(label, sensuLabel, "", "", fromDate, toDate)
			if err != nil {
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
			return event, err
		}
		for _, v := range dashboardSuggested {
			output := fmt.Sprintf("grafana_%s_url", strings.ToLower(v.GrafanaAnnotation))
			grafanaURL, err := url.Parse(v.DashboardURL)
			if err != nil {
				return event, err
			}
			if !checkMissingOrgID(grafanaURL.Query()) {
				return event, fmt.Errorf("Missing orgId in grafana URL in --grafana-dashboard-suggested. e. https://grafana.com/?orgId=1")
			}
			timeRange := fmt.Sprintf("&from=%d&to=%d", fromDate, toDate)
			finalURI := ""
			validFinalURI := false
			count := 0
			for _, s := range v.Labels {
				// &var-namespace=test
				value := ""
				value, validFinalURI = extractLabels(event, s)
				if validFinalURI {
					finalURI += fmt.Sprintf("&var-%s=%s", s, value)
					count++
				}
			}
			if validFinalURI && len(v.Labels) == count {
				annotations[output] = fmt.Sprintf("%s%s%s", grafanaURL, timeRange, finalURI)
			}
		}

	}
	// copy all annotations from event.check
	if event.Check.Annotations != nil {
		for k, v := range event.Check.Annotations {
			annotations[k] = v
		}
	}
	// add new annotations map with grafana URLs
	event.Check.Annotations = annotations
	fmt.Printf("%#v", event)
	return event, nil
}

func generateGrafanaURL(l, a, v, n string, fromDate, toDate int64) (string, error) {
	grafanaURL, err := grafanaExploreURLEncoded(l, a, v, mutatorConfig.GrafanaURL, n, mutatorConfig.GrafanaLokiDatasource, fromDate, toDate)
	if err != nil {
		return "", fmt.Errorf("Cannot generate grafana loki Explore URL")
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
	// | %7C
	// value = strings.ReplaceAll(value, "|", "%7C")
	// = %3D
	// value = strings.ReplaceAll(value, "=", "%3D")
	// space +
	// value = strings.ReplaceAll(value, " ", "+")
	return value
}

func grafanaExploreURLEncoded(label, app, value, grafana, namespace, datasource string, fromDate, toDate int64) (string, error) {
	// grafana URL expected: https://grafana.com/?orgId=1
	grafanaURL, err := url.Parse(grafana)
	if err != nil {
		return "", err
	}
	// if grafana URL not contain "?orgId=1" return a error
	if !checkMissingOrgID(grafanaURL.Query()) {
		return "", fmt.Errorf("Missing orgId in grafana URL. e. https://grafana.com/?orgId=1")
	}
	grafanaURL.Path = "explore"
	grafanaExploreURL := fmt.Sprintf("%s&left=", grafanaURL)
	searchText := url.QueryEscape(fmt.Sprintf("{%s=\\\"%s\\\"}|=\\\"%s\\\"", label, app, value))
	if namespace != "" {
		searchText = url.QueryEscape(fmt.Sprintf("{%s=\\\"%s\\\",namespace=\\\"%s\\\"}|=\\\"%s\\\"", label, app, namespace, value))
	}
	if value == "" && namespace == "" {
		searchText = url.QueryEscape(fmt.Sprintf("{%s=\\\"%s\\\"}", label, app))
	}
	grafanaExploreURI := fmt.Sprintf("[\"%d\",\"%d\",\"%s\",{\"expr\":\"%s\"}]", fromDate, toDate, datasource, searchText)
	result := fmt.Sprintf("%s%s", grafanaExploreURL, replaceSpecial(grafanaExploreURI))
	return result, nil
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
