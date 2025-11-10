// converter/converter.go
package converter

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/nektos/act/pkg/model"
	"gopkg.in/yaml.v2"
)

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
	printPlan(plan)

	// // Convert the plan to Argo Workflow
	// argoWorkflow, err := convertPlanToArgoWorkflow(plan, workflow)
	// if err != nil {
	// 	return "", fmt.Errorf("转换为 Argo Workflow 失败: %w", err)
	// }

	// // Convert to YAML
	// yamlResult, err := yaml.Marshal(argoWorkflow)
	// if err != nil {
	// 	return "", fmt.Errorf("序列化 YAML 失败: %w", err)
	// }
	var workflow wfv1.Workflow
	saveArgoWorkflowAsYAML(workflow, "converted-workflow")

	return string("converted argo workflow successfully"), nil
}

// // convertPlanToArgoWorkflow converts the plan and workflow to Argo Workflow structure
// func convertPlanToArgoWorkflow(plan *model.Plan, workflow *model.Workflow) (*ArgoWorkflow, error) {
// 	// Create the base Argo Workflow structure
// 	argoWorkflow := &ArgoWorkflow{
// 		APIVersion: "argoproj.io/v1alpha1",
// 		Kind:       "Workflow",
// 		Metadata: Metadata{
// 			GenerateName: fmt.Sprintf("%s-", workflow.Name),
// 		},
// 		Spec: WorkflowSpec{
// 			Entrypoint: workflow.Name,
// 		},
// 	}
// 	for i, stage := range plan.Stages {
// 		for j, run := range stage.Runs {
// 			log.Printf("Processing stage %d, run %d: %s", i+1, j+1, run.String())
// 			job := run.Job()
// 			template, err := convertJobToScriptTemplate(job)
// 			if err != nil {
// 				return nil, fmt.Errorf("转换 job 到脚本模板失败: %w", err)
// 			}
// 			argoWorkflow.Templates = append(argoWorkflow.Templates, *template)
// 		}
// 	}

// 	return argoWorkflow, nil
// }

// func convertJobToScriptTemplate(job *model.Job) (*Template, error) {

// }

func printPlan(plan *model.Plan) {
	fmt.Printf("执行计划包含 %d 个阶段:\n", len(plan.Stages))

	for i, stage := range plan.Stages {
		fmt.Printf("  阶段 %d: %d 个并行任务\n", i+1, len(stage.Runs))

		for _, run := range stage.Runs {
			job := run.Job()
			fmt.Printf("    - 任务: %s\n", run.String())
			fmt.Printf("      运行环境: %v\n", job.RunsOn())
			fmt.Printf("      步骤数量: %d\n", len(job.Steps))

			// 打印依赖关系
			if needs := job.Needs(); len(needs) > 0 {
				fmt.Printf("      依赖: %v\n", needs)
			}

			// 打印环境变量
			if env := job.Environment(); len(env) > 0 {
				fmt.Printf("      环境变量: %d 个\n", len(env))
			}

			// 打印 matrix 策略
			if matrix, err := job.GetMatrixes(); err == nil && len(matrix) > 1 {
				fmt.Printf("      Matrix 策略: %d 个组合\n", len(matrix))
			}
		}
	}
	fmt.Println()
}

// saveArgoWorkflowAsYAML saves the ArgoWorkflow as a YAML file in the output directory
func saveArgoWorkflowAsYAML(argoWorkflow wfv1.Workflow, workflowName string) error {
	// Create output directory if it doesn't exist
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建 output 目录失败: %w", err)
	}

	// Convert ArgoWorkflow to YAML
	yamlData, err := yaml.Marshal(argoWorkflow)
	if err != nil {
		return fmt.Errorf("序列化 YAML 失败: %w", err)
	}

	// Write to file
	filename := filepath.Join(outputDir, fmt.Sprintf("%s.yaml", workflowName))
	if err := os.WriteFile(filename, yamlData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
