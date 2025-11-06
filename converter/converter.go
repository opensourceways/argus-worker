// converter/converter.go
package converter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/nektos/act/pkg/model"
	"gopkg.in/yaml.v3"
)

// ArgoWorkflow represents the target Argo Workflow structure
type ArgoWorkflow struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   Metadata     `json:"metadata"`
	Spec       WorkflowSpec `json:"spec"`
}

type Metadata struct {
	GenerateName string            `json:"generateName,omitempty"`
	Name         string            `json:"name,omitempty"`
	Namespace    string            `json:"namespace,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

type WorkflowSpec struct {
	Entrypoint           string                `json:"entrypoint"`
	Templates            []Template            `json:"templates"`
	Arguments            Arguments             `json:"arguments,omitempty"`
	VolumeClaimTemplates []VolumeClaimTemplate `json:"volumeClaimTemplates,omitempty"`
}

type Arguments struct {
	Parameters []Parameter `json:"parameters,omitempty"`
}

type Parameter struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

type Template struct {
	Name         string            `json:"name"`
	Inputs       Inputs            `json:"inputs,omitempty"`
	Outputs      Outputs           `json:"outputs,omitempty"`
	Steps        []StepGroup       `json:"steps,omitempty"`
	DAG          *DAG              `json:"dag,omitempty"`
	Script       *Script           `json:"script,omitempty"`
	Container    *Container        `json:"container,omitempty"`
	Volumes      []Volume          `json:"volumes,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type Inputs struct {
	Parameters []Parameter `json:"parameters,omitempty"`
	Artifacts  []Artifact  `json:"artifacts,omitempty"`
}

type Outputs struct {
	Parameters []Parameter `json:"parameters,omitempty"`
	Artifacts  []Artifact  `json:"artifacts,omitempty"`
}

type Artifact struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type StepGroup []Step

type Step struct {
	Name      string                 `json:"name"`
	Template  string                 `json:"template"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	When      string                 `json:"when,omitempty"`
}

type DAG struct {
	Tasks []DAGTask `json:"tasks"`
}

type DAGTask struct {
	Name         string    `json:"name"`
	Template     string    `json:"template"`
	Dependencies []string  `json:"dependencies,omitempty"`
	Arguments    Arguments `json:"arguments,omitempty"`
	When         string    `json:"when,omitempty"`
}

type Script struct {
	Image        string        `json:"image"`
	Command      []string      `json:"command"`
	Source       string        `json:"source"`
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`
}

type Container struct {
	Image   string   `json:"image"`
	Command []string `json:"command"`
	Args    []string `json:"args,omitempty"`
}

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
}

type Volume struct {
	Name string `json:"name"`
	// Add other volume fields as needed
}

type VolumeClaimTemplate struct {
	// Define as needed
}

// ConvertWorkflow is the core conversion function that converts GitHub workflow to Argo Workflow
func ConvertWorkflow(yamlData []byte) (string, error) {
	var workflow model.Workflow

	err := yaml.Unmarshal(yamlData, &workflow)
	if err != nil {
		return "", fmt.Errorf("解析 YAML 失败: %w", err)
	}

	// Create the Argo Workflow structure
	argoWorkflow := ArgoWorkflow{
		APIVersion: "argoproj.io/v1alpha1",
		Kind:       "Workflow",
		Metadata: Metadata{
			GenerateName: fmt.Sprintf("%s-", strings.ToLower(workflow.Name)),
		},
		Spec: WorkflowSpec{
			Entrypoint: workflow.Name,
		},
	}

	// Convert jobs to templates and DAG tasks
	templates, dagTasks, err := convertJobsToTemplatesAndDAG(workflow.Jobs)
	if err != nil {
		return "", fmt.Errorf("转换 jobs 失败: %w", err)
	}

	// Add the main template that uses DAG
	mainTemplate := Template{
		Name: workflow.Name,
		DAG: &DAG{
			Tasks: dagTasks,
		},
	}

	// Add all templates to the workflow
	argoWorkflow.Spec.Templates = append([]Template{mainTemplate}, templates...)

	// Convert to JSON
	jsonData, err := json.MarshalIndent(argoWorkflow, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	return string(jsonData), nil
}

// convertJobsToTemplatesAndDAG converts GitHub jobs to Argo templates and DAG tasks
func convertJobsToTemplatesAndDAG(jobs map[string]*model.Job) ([]Template, []DAGTask, error) {
	var templates []Template
	var dagTasks []DAGTask

	// First, create all templates
	jobTemplateMap := make(map[string]Template)
	for jobID, job := range jobs {
		template, err := convertJobToTemplate(jobID, job)
		if err != nil {
			return nil, nil, fmt.Errorf("转换 job %s 为 template 失败: %w", jobID, err)
		}
		jobTemplateMap[jobID] = template
		templates = append(templates, template)
	}

	// Then, create DAG tasks with proper dependencies
	for jobID, job := range jobs {
		dagTask := DAGTask{
			Name:     jobID,
			Template: jobID,
		}

		// Convert needs to dependencies (requirement #3)
		dependencies := convertNeedsToDependencies(job)
		if len(dependencies) > 0 {
			dagTask.Dependencies = dependencies
		}

		// Convert if to when conditions (requirement #4)
		whenCondition := convertIfToWhen(job)
		if whenCondition != "" {
			dagTask.When = whenCondition
		}

		dagTasks = append(dagTasks, dagTask)
	}

	return templates, dagTasks, nil
}

// convertJobToTemplate converts a single GitHub job to an Argo template
func convertJobToTemplate(jobID string, job *model.Job) (Template, error) {
	template := Template{
		Name: jobID,
	}

	// Convert container image to script image (requirement #6)
	containerImage := "alpine:latest" // default image

	// Use reflection to access Container field
	jobValue := reflect.ValueOf(*job)
	jobType := jobValue.Type()

	for i := 0; i < jobValue.NumField(); i++ {
		field := jobValue.Field(i)
		fieldType := jobType.Field(i)

		// Look for container field
		if strings.ToLower(fieldType.Name) == "container" && field.IsValid() {
			if field.Kind() == reflect.Struct || field.Kind() == reflect.Ptr {
				containerValue := field
				if containerValue.Kind() == reflect.Ptr && !containerValue.IsNil() {
					containerValue = containerValue.Elem()
				}

				if containerValue.Kind() == reflect.Struct {
					// Look for Image field in container
					for j := 0; j < containerValue.NumField(); j++ {
						containerField := containerValue.Field(j)
						containerFieldType := containerValue.Type().Field(j)

						if strings.ToLower(containerFieldType.Name) == "image" && containerField.Kind() == reflect.String {
							if containerField.String() != "" {
								containerImage = containerField.String()
							}
							break
						}
					}
				}
			}
			break
		}
	}

	// Convert steps to merged bash script (requirements #7 and #8)
	scriptSource := convertStepsToScript(job.Steps)

	template.Script = &Script{
		Image:   containerImage,
		Command: []string{"bash"},
		Source:  scriptSource,
	}

	// Convert run-on to nodeSelector (requirement #5)
	// Note: In a full implementation, this would merge with configmap data
	nodeSelector := convertRunsOnToNodeSelector(job)
	if len(nodeSelector) > 0 {
		template.NodeSelector = nodeSelector
	}

	return template, nil
}

// convertStepsToScript merges all steps into a single bash script
func convertStepsToScript(steps []*model.Step) string {
	var scriptLines []string

	// Add a shebang and error handling
	scriptLines = append(scriptLines, "#!/bin/bash")
	scriptLines = append(scriptLines, "set -e")
	scriptLines = append(scriptLines, "")

	// Process each step
	for _, step := range steps {
		if step.Name != "" {
			scriptLines = append(scriptLines, "# "+step.Name)
		}
		if step.Run != "" {
			// Add the run command
			scriptLines = append(scriptLines, step.Run)
			scriptLines = append(scriptLines, "")
		}
		// Note: We're ignoring other step types like 'uses' for now
		// In a full implementation, we would need to handle them
	}

	return strings.Join(scriptLines, "\n")
}

// convertNeedsToDependencies converts GitHub job needs to Argo dependencies
func convertNeedsToDependencies(job *model.Job) []string {
	// Use reflection to access Needs field
	jobValue := reflect.ValueOf(*job)
	jobType := jobValue.Type()

	for i := 0; i < jobValue.NumField(); i++ {
		field := jobValue.Field(i)
		fieldType := jobType.Field(i)

		// Look for needs field
		if strings.ToLower(fieldType.Name) == "needs" && field.IsValid() {
			var dependencies []string

			switch field.Kind() {
			case reflect.String:
				if field.String() != "" {
					dependencies = append(dependencies, field.String())
				}
			case reflect.Slice:
				for j := 0; j < field.Len(); j++ {
					item := field.Index(j)
					if item.Kind() == reflect.String && item.String() != "" {
						dependencies = append(dependencies, item.String())
					}
				}
			}

			return dependencies
		}
	}

	return []string{}
}

// convertIfToWhen converts GitHub job if condition to Argo when condition
func convertIfToWhen(job *model.Job) string {
	// Use reflection to access If field
	jobValue := reflect.ValueOf(*job)
	jobType := jobValue.Type()

	for i := 0; i < jobValue.NumField(); i++ {
		field := jobValue.Field(i)
		fieldType := jobType.Field(i)

		// Look for if field
		if strings.ToLower(fieldType.Name) == "if" && field.IsValid() {
			if field.Kind() == reflect.String {
				return field.String()
			}
			break
		}
	}

	return ""
}

// convertRunsOnToNodeSelector converts GitHub job runs-on to Kubernetes nodeSelector
func convertRunsOnToNodeSelector(job *model.Job) map[string]string {
	nodeSelector := make(map[string]string)

	// Use reflection to access RunsOn field
	jobValue := reflect.ValueOf(*job)
	jobType := jobValue.Type()

	for i := 0; i < jobValue.NumField(); i++ {
		field := jobValue.Field(i)
		fieldType := jobType.Field(i)

		// Look for runs-on field (could be "RunsOn" or "RunsOn")
		if (strings.ToLower(fieldType.Name) == "runson" || strings.Contains(strings.ToLower(fieldType.Name), "runs")) && field.IsValid() {
			switch field.Kind() {
			case reflect.String:
				if field.String() != "" {
					// Simple mapping - in practice, this would be merged with configmap data
					nodeSelector["node-type"] = field.String()
				}
			case reflect.Slice:
				if field.Len() > 0 {
					item := field.Index(0)
					if item.Kind() == reflect.String && item.String() != "" {
						nodeSelector["node-label"] = item.String()
					}
				}
			case reflect.Map:
				// Handle map of labels
				for _, key := range field.MapKeys() {
					if key.Kind() == reflect.String {
						value := field.MapIndex(key)
						if value.Kind() == reflect.String {
							nodeSelector[key.String()] = value.String()
						}
					}
				}
			}
			break
		}
	}

	return nodeSelector
}
