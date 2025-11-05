package common

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetConfigMap 从 Kubernetes 集群中获取指定 namespace 和名称的 ConfigMap
// 参数:
//   - kubeconfigPath: kubeconfig 文件路径，如果为空则使用默认配置
//   - namespace: ConfigMap 所在的命名空间
//   - name: ConfigMap 的名称
//
// 返回值:
//   - *corev1.ConfigMap: 获取到的 ConfigMap 对象
//   - error: 错误信息
func GetConfigMap(kubeconfigPath, namespace, name string) (*corev1.ConfigMap, error) {
	// 获取 Kubernetes 客户端
	kubeClient, err := GetKubeClient(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes client: %w", err)
	}

	// 获取 ConfigMap
	configMap, err := kubeClient.Clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap %s in namespace %s: %w", name, namespace, err)
	}

	return configMap, nil
}

// ListConfigMaps 列出指定 namespace 中的所有 ConfigMap
// 参数:
//   - kubeconfigPath: kubeconfig 文件路径，如果为空则使用默认配置
//   - namespace: ConfigMap 所在的命名空间
//
// 返回值:
//   - *corev1.ConfigMapList: ConfigMap 列表
//   - error: 错误信息
func ListConfigMaps(kubeconfigPath, namespace string) (*corev1.ConfigMapList, error) {
	// 获取 Kubernetes 客户端
	kubeClient, err := GetKubeClient(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes client: %w", err)
	}

	// 列出 ConfigMap
	configMaps, err := kubeClient.Clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps in namespace %s: %w", namespace, err)
	}

	return configMaps, nil
}
