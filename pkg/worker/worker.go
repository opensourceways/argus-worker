package worker

import (
	"github.com/opensourceways/argus-worker/pkg/converter"
	"gopkg.in/yaml.v3"
)

// ConvertWorkflow 转换 GitHub Actions 工作流为 Argo Workflow
func ConvertWorkflow(yamlData []byte) (string, error) {
	return converter.ConvertWorkflow(yamlData)
}

func WorkerRun(yamlData []byte) (string, error) {
	var data map[string]interface{}
	err := yaml.Unmarshal(yamlData, &data)
	if err != nil {
		return "", err
	}

	convertedData, err := ConvertWorkflow(yamlData)
	if err != nil {
		return "", err
	}

	return convertedData, nil
}
