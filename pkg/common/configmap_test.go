package common

import (
	"context"
	"sync"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestGetConfigMap 测试 GetConfigMap 函数
func TestGetConfigMap(t *testing.T) {
	// 保存原始的 kubeClient 和 once
	originalKubeClient := kubeClient
	originalOnce := once

	// 确保在测试结束后恢复原始状态
	defer func() {
		kubeClient = originalKubeClient
		once = originalOnce
	}()

	// 创建一个假的 Kubernetes 客户端
	fakeClientset := fake.NewSimpleClientset()

	// 创建一个测试用的 ConfigMap
	testConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "default",
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	// 将测试 ConfigMap 添加到假客户端中
	_, err := fakeClientset.CoreV1().ConfigMaps("default").Create(context.TODO(), testConfigMap, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test configmap: %v", err)
	}

	// 替换全局 kubeClient 以使用假客户端进行测试
	kubeClient = &KubeClient{
		Clientset: fakeClientset,
	}

	// 重置 once 以确保下次调用 GetKubeClient 时使用我们的假客户端
	once = sync.Once{}

	// 测试正常情况
	configMap, err := GetConfigMap("", "default", "test-configmap")
	if err != nil {
		t.Errorf("GetConfigMap() error = %v, want nil", err)
		return
	}

	if configMap == nil {
		t.Error("GetConfigMap() returned nil configmap, want not nil")
		return
	}

	if configMap.Name != "test-configmap" {
		t.Errorf("GetConfigMap() returned configmap with name %s, want test-configmap", configMap.Name)
	}

	if configMap.Namespace != "default" {
		t.Errorf("GetConfigMap() returned configmap with namespace %s, want default", configMap.Namespace)
	}

	// 验证数据
	if len(configMap.Data) != 2 {
		t.Errorf("GetConfigMap() returned configmap with %d data items, want 2", len(configMap.Data))
	}

	if configMap.Data["key1"] != "value1" {
		t.Errorf("GetConfigMap() returned configmap with key1=%s, want value1", configMap.Data["key1"])
	}

	if configMap.Data["key2"] != "value2" {
		t.Errorf("GetConfigMap() returned configmap with key2=%s, want value2", configMap.Data["key2"])
	}

	// 测试不存在的 ConfigMap
	_, err = GetConfigMap("", "default", "non-existent-configmap")
	if err == nil {
		t.Error("GetConfigMap() for non-existent configmap should return error, got nil")
	}
}

// TestListConfigMaps 测试 ListConfigMaps 函数
func TestListConfigMaps(t *testing.T) {
	// 保存原始的 kubeClient 和 once
	originalKubeClient := kubeClient
	originalOnce := once

	// 确保在测试结束后恢复原始状态
	defer func() {
		kubeClient = originalKubeClient
		once = originalOnce
	}()

	// 创建一个假的 Kubernetes 客户端
	fakeClientset := fake.NewSimpleClientset()

	// 创建测试用的 ConfigMap
	testConfigMap1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap-1",
			Namespace: "default",
		},
		Data: map[string]string{
			"key1": "value1",
		},
	}

	testConfigMap2 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap-2",
			Namespace: "default",
		},
		Data: map[string]string{
			"key2": "value2",
		},
	}

	// 将测试 ConfigMap 添加到假客户端中
	_, err := fakeClientset.CoreV1().ConfigMaps("default").Create(context.TODO(), testConfigMap1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test configmap 1: %v", err)
	}

	_, err = fakeClientset.CoreV1().ConfigMaps("default").Create(context.TODO(), testConfigMap2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test configmap 2: %v", err)
	}

	// 替换全局 kubeClient 以使用假客户端进行测试
	kubeClient = &KubeClient{
		Clientset: fakeClientset,
	}

	// 重置 once 以确保下次调用 GetKubeClient 时使用我们的假客户端
	once = sync.Once{}

	// 测试列出 ConfigMap
	configMaps, err := ListConfigMaps("", "default")
	if err != nil {
		t.Errorf("ListConfigMaps() error = %v, want nil", err)
		return
	}

	if configMaps == nil {
		t.Error("ListConfigMaps() returned nil configmaps, want not nil")
		return
	}

	if len(configMaps.Items) != 2 {
		t.Errorf("ListConfigMaps() returned %d configmaps, want 2", len(configMaps.Items))
	}

	// 验证返回的 ConfigMap
	foundConfigMap1 := false
	foundConfigMap2 := false

	for _, cm := range configMaps.Items {
		if cm.Name == "test-configmap-1" && cm.Namespace == "default" {
			foundConfigMap1 = true
			if cm.Data["key1"] != "value1" {
				t.Errorf("ConfigMap 1 has incorrect data for key1: %s, want value1", cm.Data["key1"])
			}
		}
		if cm.Name == "test-configmap-2" && cm.Namespace == "default" {
			foundConfigMap2 = true
			if cm.Data["key2"] != "value2" {
				t.Errorf("ConfigMap 2 has incorrect data for key2: %s, want value2", cm.Data["key2"])
			}
		}
	}

	if !foundConfigMap1 {
		t.Error("ListConfigMaps() did not return test-configmap-1")
	}

	if !foundConfigMap2 {
		t.Error("ListConfigMaps() did not return test-configmap-2")
	}

	// 测试空 namespace 的情况
	emptyConfigMaps, err := ListConfigMaps("", "empty-namespace")
	if err != nil {
		t.Errorf("ListConfigMaps() for empty namespace error = %v, want nil", err)
		return
	}

	if len(emptyConfigMaps.Items) != 0 {
		t.Errorf("ListConfigMaps() for empty namespace returned %d configmaps, want 0", len(emptyConfigMaps.Items))
	}
}
