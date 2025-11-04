// converter/converter.go
package converter

import (
	"encoding/json"
	"fmt"

	"github.com/nektos/act/pkg/model"
	"gopkg.in/yaml.v3"
)

// SimplifiedWorkflow 是我们“转换”后的目标结构
type SimplifiedWorkflow struct {
	Name string   `json:"name"`
	Jobs []string `json:"jobs"`
}

// ConvertWorkflow 是核心的转换函数
func ConvertWorkflow(yamlData []byte) (string, error) {
	var workflow model.Workflow

	err := yaml.Unmarshal(yamlData, &workflow)
	if err != nil {
		return "", fmt.Errorf("解析 YAML 失败: %w", err)
	}

	jobIDs := make([]string, 0, len(workflow.Jobs))
	for jobID := range workflow.Jobs {
		jobIDs = append(jobIDs, jobID)
	}

	simplified := SimplifiedWorkflow{
		Name: workflow.Name,
		Jobs: jobIDs,
	}

	jsonData, err := json.Marshal(simplified)
	if err != nil {
		return "", fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	return string(jsonData), nil
}
