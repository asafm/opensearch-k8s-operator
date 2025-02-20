package services

import (
	"encoding/json"
	"opensearch.opster.io/opensearch-gateway/responses"
	"opensearch.opster.io/pkg/helpers"
	"strings"
)

var ClusterSettingsExcludeBrokenPath = []string{"cluster", "routing", "allocation", "exclude", "_name"}

func HasIndicesWithNoReplica(service *OsClusterClient) (bool, error) {
	response, err := service.CatIndices()
	if err != nil {
		return false, err
	}
	for _, index := range response {
		if index.Rep == "" || index.Rep == "0" {
			return true, err
		}
	}
	return false, err
}

func HasShardsOnNode(service *OsClusterClient, nodeName string) (bool, error) {
	var headers []string
	response, err := service.CatShards(headers)
	if err != nil {
		return false, err
	}
	for _, shardsData := range response {
		if shardsData.NodeName == nodeName {
			return true, err
		}
	}
	return false, err
}

func AppendExcludeNodeHost(service *OsClusterClient, nodeNameToExclude string) (bool, error) {
	response, err := service.GetClusterSettings()
	if err != nil {
		return false, err
	}
	val, ok := helpers.FindByPath(response.Transient, ClusterSettingsExcludeBrokenPath)
	var valAsString = nodeNameToExclude
	if ok && val != "" {
		valAsString = val.(string) + "," + nodeNameToExclude
	}
	settings := createClusterSettingsResponseWithExcludeName(valAsString)
	settingsAsJson, err := json.Marshal(settings)
	if err == nil {
		_, err = service.PutClusterSettings(string(settingsAsJson))
	}
	return err == nil, err
}

func RemoveExcludeNodeHost(service *OsClusterClient, nodeNameToExclude string) (bool, error) {
	response, err := service.GetClusterSettings()
	if err != nil {
		return false, err
	}
	val, ok := helpers.FindByPath(response.Transient, ClusterSettingsExcludeBrokenPath)
	if !ok || val == "" {
		return true, err
	}
	valAsString := strings.ReplaceAll(val.(string), nodeNameToExclude, "")
	valAsString = strings.ReplaceAll(valAsString, ",,", ",")
	settings := createClusterSettingsResponseWithExcludeName(valAsString)
	settingsAsJson, err := json.Marshal(settings)
	if err == nil {
		_, err = service.PutClusterSettings(string(settingsAsJson))
	}
	return err == nil, err
}

func createClusterSettingsResponseWithExcludeName(exclude string) responses.ClusterSettingsResponse {
	var val *string = nil
	if exclude != "" {
		val = &exclude
	}
	return responses.ClusterSettingsResponse{Transient: map[string]interface{}{
		"cluster": map[string]interface{}{
			"routing": map[string]interface{}{
				"allocation": map[string]interface{}{
					"exclude": map[string]interface{}{
						"_name": val,
					},
				},
			},
		},
	}}
}
