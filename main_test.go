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
	_, err1 := grafanaExplorerURLEncoded("app", "eventrouter", "test", test1, 1606487400000, 1606487700000)
	assert.NoError(t, err1)
	test2 := "https://grafana.com/"
	_, err2 := grafanaExplorerURLEncoded("app", "eventrouter", "test", test2, 1606487400000, 1606487700000)
	assert.Error(t, err2)
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
