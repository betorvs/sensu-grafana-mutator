package main

import (
	"net/url"
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
	mutatorConfig.GrafanaExploreLinkEnabled = true
	err2 := checkArgs(event2)
	assert.NoError(err2)
}

func TestGrafanaExploreURLEncoded(t *testing.T) {
	test1 := "https://grafana.com/?orgId=1"
	expected1 := "%22%7Bapp%3D%5C%22eventrouter%5C%22%7D%7C%3D%5C%22test"
	result1, err1 := grafanaExploreURLEncoded("app", "eventrouter", "test", test1, "", "loki", 1606487400000, 1606487700000)
	assert.NoError(t, err1)
	assert.Contains(t, result1, expected1)
	test2 := "https://grafana.com/"
	_, err2 := grafanaExploreURLEncoded("app", "eventrouter", "test", test2, "", "loki", 1606487400000, 1606487700000)
	assert.Error(t, err2)
	namespace := "spacename"
	result3, err3 := grafanaExploreURLEncoded("app", "eventrouter", "test", test1, namespace, "loki", 1606487400000, 1606487700000)
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

func TestCheckMissingOrgID(t *testing.T) {
	grafanaURL1, _ := url.Parse("https://grafana.com/?orgId=1")
	value1 := checkMissingOrgID(grafanaURL1.Query())
	assert.True(t, value1)
	grafanaURL2, _ := url.Parse("https://grafana.com/?orgid=1")
	value2 := checkMissingOrgID(grafanaURL2.Query())
	assert.False(t, value2)
}

func TestExtractLabels(t *testing.T) {
	event1 := v2.FixtureEvent("entity1", "check1")
	event1.Labels["test1"] = "value1"
	value1, result1 := extractLabels(event1, "test1")
	assert.Contains(t, value1, "value1")
	assert.True(t, result1)
	event2 := v2.FixtureEvent("entity2", "check2")
	_, result2 := extractLabels(event2, "test2")
	assert.False(t, result2)
}
