package main

import (
	"testing"

	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	event := v2.FixtureEvent("entity1", "check1")
	err := checkArgs(event)
	assert.Error(err)
	event2 := v2.FixtureEvent("entity2", "check2")
	mutatorConfig.GrafanaURL = "http://127.0.0.1:3000/?orgId=1"
	mutatorConfig.GrafanaLokiExplorerPipeline = "k8s_id"
	err2 := checkArgs(event2)
	assert.NoError(err2)
}

func TestGrafanaExplorerURLEncoded(t *testing.T) {
	test1 := "https://grafana.com/?orgId=1"
	expected1 := "%22%7Bapp%3D%5C%22eventrouter%5C%22%7D%7C%3D%5C%22test"
	result1, err1 := grafanaExplorerURLEncoded("app", "eventrouter", "test", test1, "", "loki", 1606487400000, 1606487700000)
	assert.NoError(t, err1)
	assert.Contains(t, result1, expected1)
	test2 := "https://grafana.com/"
	_, err2 := grafanaExplorerURLEncoded("app", "eventrouter", "test", test2, "", "loki", 1606487400000, 1606487700000)
	assert.Error(t, err2)
	namespace := "spacename"
	result3, err3 := grafanaExplorerURLEncoded("app", "eventrouter", "test", test1, namespace, "loki", 1606487400000, 1606487700000)
	assert.NoError(t, err3)
	assert.Contains(t, result3, namespace)
}

func TestReplaceSpecial(t *testing.T) {
	test1 := "ads[]{}\""
	expected1 := "ads%5B%5D%7B%7D%22"
	result1 := replaceSpecial(test1)
	assert.NotContains(t, result1, "[")
	assert.NotContains(t, result1, "]")
	assert.NotContains(t, result1, "{")
	assert.NotContains(t, result1, "}")
	assert.NotContains(t, result1, "\"")
	assert.Equal(t, result1, expected1)
}
