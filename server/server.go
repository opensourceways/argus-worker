// server/server.go
package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/opensourceways/argus-worker/converter" // 替换为你的实际 module 名称

	"github.com/gin-gonic/gin"
)

// ConversionJob 定义任务
type ConversionJob struct {
	Payload    []byte
	ResultChan chan ConversionResult
}

type ConversionResult struct {
	Data  string
	Error error
}

var JobQueue chan ConversionJob

const (
	MaxWorkers = 5
	MaxQueue   = 100
)

// StartWorkerPool 启动工作池
func StartWorkerPool() {
	JobQueue = make(chan ConversionJob, MaxQueue)

	for i := 1; i <= MaxWorkers; i++ {
		go func(workerID int) {
			log.Printf("Worker %d 启动", workerID)
			for job := range JobQueue {
				log.Printf("Worker %d 开始处理任务", workerID)
				convertedData, err := converter.ConvertWorkflow(job.Payload)
				job.ResultChan <- ConversionResult{
					Data:  convertedData,
					Error: err,
				}
			}
		}(i)
	}
}

// handleConversion Gin 处理器
func handleConversion(c *gin.Context) {
	// 只允许 POST
	if c.Request.Method != http.MethodPost {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
			"error": "只允许 POST 方法",
		})
		return
	}

	// 读取 body
	body, err := c.GetRawData()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "读取请求体失败",
		})
		return
	}

	if len(body) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "请求体为空",
		})
		return
	}

	resultChan := make(chan ConversionResult)
	job := ConversionJob{
		Payload:    body,
		ResultChan: resultChan,
	}

	select {
	case JobQueue <- job:
		log.Println("任务已提交到队列")
	default:
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"error": "服务繁忙，任务队列已满",
		})
		return
	}

	log.Println("等待任务结果...")
	result := <-resultChan

	if result.Error != nil {
		log.Printf("任务处理失败: %v", result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("转换失败: %v", result.Error),
		})
		return
	}

	log.Println("任务处理成功")
	c.Data(http.StatusOK, "application/json", []byte(result.Data))
}

// NewRouter 创建 Gin 路由
func NewRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/api/v1/convert", handleConversion)

	return r
}
