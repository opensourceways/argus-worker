package converter

import (
	"fmt"

	"github.com/opensourceways/argus-worker/pkg/common"
)

func (c *WorkflowConverter) parseRunsOn(runsOn []string) string {
	configMapName := runsOn[0]
	configMap, err := common.GetConfigMap("/home/k9s/argo/openmerlin-guiyang-006-cluster-kubeconfig.yaml", "argo", configMapName) // 假设 ConfigMap 名称为 "runsOn"，命名空间为 "default"
	if err != nil {
		fmt.Printf("Warning: Failed to get ConfigMap 'runsOn': %v\n", err)
		// 继续使用默认的硬编码映射
	} else {
		key := configMapName + ".yaml"
		if value, exists := configMap.Data[key]; exists {
			return value
		}
	}

	return ""
}
