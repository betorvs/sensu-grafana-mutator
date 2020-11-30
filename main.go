package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// Config represents the mutator plugin config.
type Config struct {
	sensu.PluginConfig
	GrafanaURL                        string
	GrafanaLokiExplorerStreamLabel    string
	GrafanaLokiExplorerStreamSelector string
	GrafanaLokiExplorerPipeline       string
	GrafanaLokiExplorerRange          int
	TimeRange                         int64
}

var (
	mutatorConfig = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-grafana-mutator",
			Short:    "Sensu grafana mutator add grafana_url annotation",
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
			Path:      "grafana-loki-explorer-stream-label",
			Env:       "GRAFANA_LOKI_EXPLORER_STREAM_LABEL",
			Argument:  "grafana-loki-explorer-stream-label",
			Shorthand: "l",
			Default:   "app",
			Usage:     "From Grafana Loki streams use label. e. {app=eventrouter} then '-l app' ",
			Value:     &mutatorConfig.GrafanaLokiExplorerStreamLabel,
		},
		{
			Path:      "grafana-loki-explorer-stream-selector",
			Env:       "GRAFANA_LOKI_EXPLORER_STREAM_SELECTOR",
			Argument:  "grafana-loki-explorer-stream-selector",
			Shorthand: "s",
			Default:   "eventrouter",
			Usage:     "From Grafana Loki streams use label. e. {app=eventrouter} then '-s eventrouter' ",
			Value:     &mutatorConfig.GrafanaLokiExplorerStreamSelector,
		},
		{
			Path:      "grafana-loki-explorer-pipeline",
			Env:       "GRAFANA_LOKI_EXPLORER_PIPELINE",
			Argument:  "grafana-loki-explorer-pipeline",
			Shorthand: "p",
			Default:   "",
			Usage:     "From Sensu Events, choose one label to be parse here. e. {app=eventrouter} |= k8s_id then use -p k8s_id",
			Value:     &mutatorConfig.GrafanaLokiExplorerPipeline,
		},
		{
			Path:      "grafana-loki-explorer-range",
			Env:       "",
			Argument:  "grafana-loki-explorer-range",
			Shorthand: "r",
			Default:   300,
			Usage:     "Time range in seconds to create grafana explorer URL",
			Value:     &mutatorConfig.GrafanaLokiExplorerRange,
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
	if mutatorConfig.GrafanaLokiExplorerPipeline == "" {
		return fmt.Errorf("--grafana-loki-explorer-pipeline or GRAFANA_LOKI_EXPLORER_PIPELINE environment variable is required")
	}
	mutatorConfig.TimeRange = int64(mutatorConfig.GrafanaLokiExplorerRange * 1000)
	return nil
}

func executeMutator(event *types.Event) (*types.Event, error) {
	log.Println("executing mutator with --grafana-url", mutatorConfig.GrafanaURL)
	if mutatorConfig.GrafanaLokiExplorerPipeline != "" {
		annotations := make(map[string]string)
		fromDate := event.Timestamp * 1000
		toDate := event.Timestamp*1000 + mutatorConfig.TimeRange
		if event.Labels != nil {
			for k, v := range event.Labels {
				if k == mutatorConfig.GrafanaLokiExplorerPipeline {
					grafanaURL, err := generateGrafanaURL(v, fromDate, toDate)
					if err != nil {
						return event, err
					}
					annotations["grafana_loki_url"] = grafanaURL
				}
			}
		}
		if event.Entity.Labels != nil {
			for k, v := range event.Entity.Labels {
				if k == mutatorConfig.GrafanaLokiExplorerPipeline {
					grafanaURL, err := generateGrafanaURL(v, fromDate, toDate)
					if err != nil {
						return event, err
					}
					annotations["grafana_loki_url"] = grafanaURL
				}
			}
		}
		if event.Check.Labels != nil {
			for k, v := range event.Check.Labels {
				if k == mutatorConfig.GrafanaLokiExplorerPipeline {
					grafanaURL, err := generateGrafanaURL(v, fromDate, toDate)
					if err != nil {
						return event, err
					}
					annotations["grafana_loki_url"] = grafanaURL
				}
			}
		}
		event.Check.Annotations = annotations
	}
	return event, nil
}

func generateGrafanaURL(v string, fromDate, toDate int64) (string, error) {
	grafanaURL, err := grafanaExplorerURLEncoded(mutatorConfig.GrafanaLokiExplorerStreamLabel, mutatorConfig.GrafanaLokiExplorerStreamSelector, v, mutatorConfig.GrafanaURL, fromDate, toDate)
	if err != nil {
		return "", fmt.Errorf("Cannot generate grafana loki explorer URL")
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

func grafanaExplorerURLEncoded(label, app, value, grafana string, fromDate, toDate int64) (string, error) {
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
	grafanaExplorerURL := fmt.Sprintf("%s&left=", grafanaURL)
	searchText := url.QueryEscape(fmt.Sprintf("{%s=\\\"%s\\\"}|=\\\"%s\\\"", label, app, value))
	grafanaExplorerURI := fmt.Sprintf("[\"%d\",\"%d\",\"loki\",{\"expr\":\"%s\"}]", fromDate, toDate, searchText)
	result := fmt.Sprintf("%s%s", grafanaExplorerURL, replaceSpecial(grafanaExplorerURI))
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
