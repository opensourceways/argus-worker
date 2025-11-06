package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/opensourceways/argus-worker/common"
)

func main() {
	// Kubernetes 配置文件路径
	kubeconfigPath := "/home/k9s/argo/openmerlin-guiyang-006-cluster-kubeconfig-small.yaml"

	// 要读取的 ConfigMap 名称和命名空间
	namespace := "argo" // 根据实际情况修改命名空间
	configMapName := "linux-aarch64-a2b4-1"

	// 读取指定的 ConfigMap
	configMap, err := common.GetConfigMap(kubeconfigPath, namespace, configMapName)
	if err != nil {
		log.Fatalf("Failed to get ConfigMap: %v", err)
	}

	// 输出 ConfigMap 信息
	fmt.Printf("ConfigMap Name: %s\n", configMap.Name)
	fmt.Printf("ConfigMap Namespace: %s\n", configMap.Namespace)
	fmt.Printf("ConfigMap Data:\n")

	// 输出 ConfigMap 中的所有数据
	// for key, value := range configMap.Data {
	// fmt.Printf(" key %s: %s\n", key, value)
	if specValue, exists := configMap.Data["linux-aarch64-a2b4-1.yaml"]; exists {
		jsonSpecValue, err := json.MarshalIndent(specValue, "", "  ")
		if err != nil {
			log.Printf("Failed to convert specValue to JSON: %v", err)
		} else {
			fmt.Printf("Spec 内容 (JSON 格式):\n%s\n", string(jsonSpecValue))
		}
	}
	// }

	// 如果需要，也可以列出命名空间中的所有 ConfigMap
	fmt.Println("\nListing all ConfigMaps in namespace:")
	configMaps, err := common.ListConfigMaps(kubeconfigPath, namespace)
	if err != nil {
		log.Printf("Failed to list ConfigMaps: %v", err)
	} else {
		for _, cm := range configMaps.Items {
			fmt.Printf("  - %s\n", cm.Name)
		}
	}
}
