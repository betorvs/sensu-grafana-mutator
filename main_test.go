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
	test1map := map[string]string{"app": "eventrouter", "eventID": "test"}
	test1 := "https://grafana.com/?orgId=1"
	expected1 := "app%3D%5C%22eventrouter%5C%22%7D%7C%3D%5C%22test"
	result1, err1 := grafanaExploreURLEncoded(test1map, test1, "loki", 1606487400000, 1606487700000)
	assert.NoError(t, err1)
	assert.Contains(t, result1, expected1)
	test2map := map[string]string{"app": "eventrouter", "eventID": "test"}
	test2 := "https://grafana.com/"
	_, err2 := grafanaExploreURLEncoded(test2map, test2, "loki", 1606487400000, 1606487700000)
	assert.Error(t, err2)
	test3map := map[string]string{"namespace": "spacename"}
	namespace := "spacename"
	result3, err3 := grafanaExploreURLEncoded(test3map, test1, "loki", 1606487400000, 1606487700000)
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

func TestGenerateURIBySlice(t *testing.T) {
	labels := []string{"testa", "testb"}
	event1 := v2.FixtureEvent("entity1", "check1")
	event1.Labels["testa"] = "valuea"
	event1.Labels["testb"] = "valueb"
	expected1 := "&var-testa=valuea&var-testb=valueb"
	result1, res1 := generateURIBySlice(event1, labels)
	assert.True(t, res1)
	assert.Contains(t, result1, expected1)
	event2 := v2.FixtureEvent("entity2", "check2")
	event2.Labels["testa"] = "valuea"
	event2.Labels["testb"] = "valueb"
	_, res2 := generateURIBySlice(event1, labels)
	assert.True(t, res2)
}

func TestSearchMatchLabels(t *testing.T) {
	event1 := v2.FixtureEvent("entity1", "check1")
	event1.Labels["testa"] = "valuea"
	event1.Labels["testb"] = "valueb"
	event1.Labels["testc"] = "valuec"
	labels := make(map[string]string)
	res1 := searchMatchLabels(event1, labels)
	assert.False(t, res1)

	labels["testa"] = "valuea"
	labels["testc"] = "valuec"
	res2 := searchMatchLabels(event1, labels)
	assert.True(t, res2)

}

func TestMergeStringMaps(t *testing.T) {
	left1 := map[string]string{"left1": "leftValue1"}
	right1 := map[string]string{"right1": "rightValue1"}
	val1 := map[string]string{"left1": "leftValue1", "right1": "rightValue1"}
	res1 := mergeStringMaps(left1, right1)
	assert.Equal(t, val1, res1)
	left2 := map[string]string{"left1": "leftValue1"}
	right2 := map[string]string{"right1": "rightValue1", "left1": "rightValueLeft1"}
	val2 := map[string]string{"left1": "leftValue1", "right1": "rightValue1"}
	res2 := mergeStringMaps(left2, right2)
	assert.Equal(t, val2, res2)
	left3 := map[string]string{"left1": "leftValue1"}
	right3 := map[string]string{}
	val3 := map[string]string{"left1": "leftValue1"}
	res3 := mergeStringMaps(left3, right3)
	assert.Equal(t, val3, res3)
	left4 := map[string]string{}
	right4 := map[string]string{"right1": "rightValue1"}
	val4 := map[string]string{"right1": "rightValue1"}
	res4 := mergeStringMaps(left4, right4)
	assert.Equal(t, val4, res4)
}

func TestStringToSliceStrings(t *testing.T) {
	test1 := "test1"
	expected1 := []string{"test1"}
	res1 := stringToSliceStrings(test1)
	assert.Equal(t, expected1, res1)
	test2 := "test1,test2"
	expected2 := []string{"test1", "test2"}
	res2 := stringToSliceStrings(test2)
	assert.Equal(t, expected2, res2)
	test3 := "test1,"
	expected3 := []string{"test1"}
	res3 := stringToSliceStrings(test3)
	assert.Equal(t, expected3, res3)
	test4 := ""
	expected4 := []string{}
	res4 := stringToSliceStrings(test4)
	assert.Equal(t, expected4, res4)
}
