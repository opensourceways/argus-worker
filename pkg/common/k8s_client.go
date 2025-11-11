package common

import (
	"fmt"
	"os"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeClient *KubeClient
	once       sync.Once
)

// KubeClient 封装了Kubernetes客户端的结构体
type KubeClient struct {
	Clientset kubernetes.Interface
}

// GetKubeClient 获取Kubernetes客户端单例实例
// 如果kubeconfigPath为空，则尝试从环境变量或默认位置获取
func GetKubeClient(kubeconfigPath string) (*KubeClient, error) {
	var err error

	once.Do(func() {
		kubeClient, err = newKubeClient(kubeconfigPath)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return kubeClient, nil
}

// newKubeClient 创建新的Kubernetes客户端实例
func newKubeClient(kubeconfigPath string) (*KubeClient, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		// 使用指定的kubeconfig文件
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig file: %w", err)
		}
	} else {
		// 尝试从环境变量获取kubeconfig路径
		if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfigEnv)
			if err != nil {
				return nil, fmt.Errorf("failed to build config from KUBECONFIG env: %w", err)
			}
		} else {
			// 尝试使用in-cluster配置（在Pod内运行时）
			config, err = rest.InClusterConfig()
			if err != nil {
				// 如果不是in-cluster，则使用默认的kubeconfig路径
				homeDir := os.Getenv("HOME")
				if homeDir == "" {
					homeDir = os.Getenv("USERPROFILE") // Windows support
				}
				defaultKubeconfig := fmt.Sprintf("%s/.kube/config", homeDir)
				config, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
				if err != nil {
					return nil, fmt.Errorf("failed to build config from default location: %w", err)
				}
			}
		}
	}

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &KubeClient{
		Clientset: clientset,
	}, nil
}

// GetClientset 获取底层的clientset
func (kc *KubeClient) GetClientset() kubernetes.Interface {
	return kc.Clientset
}

// Reset 重置单例实例（主要用于测试）
func Reset() {
	once = sync.Once{}
	kubeClient = nil
}
