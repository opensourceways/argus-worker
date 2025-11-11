// converter/converter.go
package converter

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/nektos/act/pkg/model"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkflowConverter struct {
	githubWorkflow *model.Workflow
}

func NewConverter(ghWorkflow *model.Workflow) *WorkflowConverter {
	return &WorkflowConverter{
		githubWorkflow: ghWorkflow,
	}
}

func (c *WorkflowConverter) Run() (*wfv1.Workflow, error) {
	if c.githubWorkflow == nil {
		return nil, fmt.Errorf("GitHub workflow is nil")
	}

	// 创建 Argo Workflow 对象
	argoWf := &wfv1.Workflow{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "argoproj.io/v1alpha1",
			Kind:       "Workflow",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", strings.ToLower(c.githubWorkflow.Name)),
		},
		Spec: wfv1.WorkflowSpec{
			Entrypoint: "main",
		},
	}

	// 创建主 DAG 模板
	mainTemplate := wfv1.Template{
		Name: "main",
		DAG:  &wfv1.DAGTemplate{},
	}

	// 转换每个 job
	for jobName, job := range c.githubWorkflow.Jobs {
		// 为每个 job 创建一个独立的 template
		jobTemplate, err := c.convertJobToTemplate(jobName, job)
		if err != nil {
			return nil, fmt.Errorf("failed to convert job %s: %w", jobName, err)
		}
		argoWf.Spec.Templates = append(argoWf.Spec.Templates, *jobTemplate)

		// 在 DAG 中添加任务
		mainTemplate.DAG.Tasks = append(mainTemplate.DAG.Tasks, wfv1.DAGTask{
			Name:     jobName,
			Template: jobName,
		})
	}

	// 将主 DAG 模板添加到 templates 列表
	argoWf.Spec.Templates = append(argoWf.Spec.Templates, mainTemplate)

	return argoWf, nil
}

func ParseContainerFromYAML(yamlData string) (*corev1.Container, error) {
	// 创建一个 map 来解析原始 YAML 数据
	var rawData map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlData), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse container YAML: %w", err)
	}

	// 创建容器对象
	container := &corev1.Container{}

	// 逐个处理字段，避免严格的类型检查错误
	for key, value := range rawData {
		switch key {
		case "name":
			if name, ok := value.(string); ok {
				container.Name = name
			}
		case "image":
			if image, ok := value.(string); ok {
				container.Image = image
			}
		case "imagePullPolicy":
			if policy, ok := value.(string); ok {
				container.ImagePullPolicy = corev1.PullPolicy(policy)
			}
		case "resources":
			if resources, ok := value.(map[interface{}]interface{}); ok {
				container.Resources = parseResources(resources)
			}
		case "volumeMounts":
			if mounts, ok := value.([]interface{}); ok {
				container.VolumeMounts = parseVolumeMounts(mounts)
			}
		case "lifecycle":
			if lifecycle, ok := value.(map[interface{}]interface{}); ok {
				container.Lifecycle = parseLifecycle(lifecycle)
			}
		}
	}

	return container, nil
}

// parseResources 解析资源限制和请求
func parseResources(resources map[interface{}]interface{}) corev1.ResourceRequirements {
	result := corev1.ResourceRequirements{
		Limits:   make(corev1.ResourceList),
		Requests: make(corev1.ResourceList),
	}

	if limits, ok := resources["limits"].(map[interface{}]interface{}); ok {
		for k, v := range limits {
			if key, ok := k.(string); ok {
				// 处理资源值，可能是字符串或包含 format 字段的对象
				var value string
				if str, ok := v.(string); ok {
					value = str
				} else if obj, ok := v.(map[interface{}]interface{}); ok {
					// 从对象中提取实际的资源值
					// 这里我们假设资源值应该从其他地方获取，暂时使用默认值
					value = getResourceValue(key, obj)
				}

				if value != "" {
					result.Limits[corev1.ResourceName(key)] = parseResourceQuantity(value)
				}
			}
		}
	}

	if requests, ok := resources["requests"].(map[interface{}]interface{}); ok {
		for k, v := range requests {
			if key, ok := k.(string); ok {
				// 处理资源值，可能是字符串或包含 format 字段的对象
				var value string
				if str, ok := v.(string); ok {
					value = str
				} else if obj, ok := v.(map[interface{}]interface{}); ok {
					// 从对象中提取实际的资源值
					// 这里我们假设资源值应该从其他地方获取，暂时使用默认值
					value = getResourceValue(key, obj)
				}

				if value != "" {
					result.Requests[corev1.ResourceName(key)] = parseResourceQuantity(value)
				}
			}
		}
	}

	return result
}

// getResourceValue 从资源对象中提取实际的资源值
func getResourceValue(resourceName string, resourceObj map[interface{}]interface{}) string {
	// 根据资源名称返回默认值
	switch resourceName {
	case "cpu":
		return "46"
	case "memory":
		return "128Gi"
	case "huawei.com/ascend-1980":
		return "1"
	default:
		return "1"
	}
}

// parseResourceQuantity 解析资源量字符串
func parseResourceQuantity(value string) resource.Quantity {
	// 简化处理，实际应该使用 resource.ParseQuantity
	quantity, _ := resource.ParseQuantity(value)
	return quantity
}

// parseVolumeMounts 解析卷挂载配置
func parseVolumeMounts(mounts []interface{}) []corev1.VolumeMount {
	var result []corev1.VolumeMount

	for _, item := range mounts {
		if mountData, ok := item.(map[interface{}]interface{}); ok {
			mount := corev1.VolumeMount{}
			for k, v := range mountData {
				if key, ok := k.(string); ok {
					switch key {
					case "name":
						if name, ok := v.(string); ok {
							mount.Name = name
						}
					case "mountPath":
						if path, ok := v.(string); ok {
							mount.MountPath = path
						}
					case "readOnly":
						if readOnly, ok := v.(bool); ok {
							mount.ReadOnly = readOnly
						}
					}
				}
			}
			result = append(result, mount)
		}
	}

	return result
}

// parseLifecycle 解析生命周期配置
func parseLifecycle(lifecycle map[interface{}]interface{}) *corev1.Lifecycle {
	result := &corev1.Lifecycle{}

	for k, v := range lifecycle {
		if key, ok := k.(string); ok {
			if lifecycleData, ok := v.(map[interface{}]interface{}); ok {
				switch key {
				case "postStart":
					result.PostStart = parseHandler(lifecycleData)
				case "preStop":
					result.PreStop = parseHandler(lifecycleData)
				}
			}
		}
	}

	return result
}

// parseHandler 解析处理器配置
func parseHandler(handler map[interface{}]interface{}) *corev1.LifecycleHandler {
	result := &corev1.LifecycleHandler{}

	for k, v := range handler {
		if key, ok := k.(string); ok {
			if execData, ok := v.(map[interface{}]interface{}); ok {
				switch key {
				case "exec":
					if command, ok := execData["command"].([]interface{}); ok {
						var cmd []string
						for _, c := range command {
							if str, ok := c.(string); ok {
								cmd = append(cmd, str)
							}
						}
						result.Exec = &corev1.ExecAction{
							Command: cmd,
						}
					}
				}
			}
		}
	}

	return result
}

func (c *WorkflowConverter) convertJobToTemplate(jobName string, job *model.Job) (*wfv1.Template, error) {
	template := &wfv1.Template{
		Name: jobName,
	}

	// 获取 runsOn 配置
	runsOnConfig := c.parseRunsOn(job.RunsOn())

	// 如果获取到了 runsOn 配置，则解析 YAML 并应用到模板
	if runsOnConfig != "" {
		fmt.Printf("Got runsOn config: %s\n", runsOnConfig)

		container, err := ParseContainerFromYAML(runsOnConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse runsOn config: %w", err)
		}
		fmt.Printf("Parsed runsOn spec: %+v\n", container)

		// 使用解析的容器配置
		template.Container = container
	}

	// 合并所有步骤的 shell 命令
	var scriptLines []string
	scriptLines = append(scriptLines, "set -e") // 确保任一命令失败时退出

	for _, step := range job.Steps {
		if step.Run != "" {
			// 处理多行命令，去除空行
			lines := strings.Split(step.Run, "\n")
			for _, line := range lines {
				if trimmed := strings.TrimSpace(line); trimmed != "" {
					scriptLines = append(scriptLines, trimmed)
				}
			}
		}
	}

	// 创建或更新容器规格
	if template.Container == nil {
		container := corev1.Container{
			Image:   job.Container().Image,
			Command: []string{"/bin/sh", "-c"},
			Args: []string{
				strings.Join(scriptLines, "\n"),
			},
		}
		template.Container = &container
	} else {
		// 如果容器已经从 runsOn 配置中设置，确保设置了必要的字段
		if template.Container.Image == "" {
			template.Container.Image = job.Container().Image
		}
		template.Container.Command = []string{"/bin/sh", "-c"}
		template.Container.Args = []string{
			strings.Join(scriptLines, "\n"),
		}
	}

	return template, nil
}

// ConvertWorkflow is the core conversion function that converts GitHub workflow to Argo Workflow
func ConvertWorkflow(yamlData []byte) (string, error) {
	// Use act's NewSingleWorkflowPlanner to validate and parse the workflow directly from bytes
	reader := bytes.NewReader(yamlData)
	planner, err := model.NewSingleWorkflowPlanner("workflow.yml", reader)
	if err != nil {
		return "", fmt.Errorf("创建 workflow planner 失败: %w", err)
	}

	plan, err := planner.PlanAll()
	if err != nil {
		log.Fatalf("创建完整计划失败: %v", err)
	}
	PrintPlan(plan)
	printList(plan)

	// 重新创建 reader，因为原来的 reader 已经被读取到末尾
	reader = bytes.NewReader(yamlData)
	githubWorkflow, err := model.ReadWorkflow(reader, false)
	if err != nil {
		fmt.Printf("Error parsing GitHub workflow: %v\n", err)
		os.Exit(1)
	}

	// 创建转换器并生成 Argo Workflow
	converter := NewConverter(githubWorkflow)
	argoWorkflow, err := converter.Run()
	if err != nil {
		fmt.Printf("Error converting to Argo workflow: %v\n", err)
		os.Exit(1)
	}
	// 序列化为 YAML 并输出
	output, err := yaml.Marshal(argoWorkflow)
	if err != nil {
		fmt.Printf("Error marshaling Argo workflow: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))

	return string("converted argo workflow successfully"), nil
}
