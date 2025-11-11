// main.go
package main

import (
	"log"

	"github.com/opensourceways/argus-worker/pkg/server" // 替换为你的实际 module 名称
)

// Run 启动应用
func Run() error {
	// 启动工作池
	server.StartWorkerPool()
	log.Println("Worker 池已启动")

	// 创建并启动 Gin 服务
	router := server.NewRouter()
	log.Println("Web 服务启动于 http://localhost:8080")
	return router.Run(":8080")
}

func main() {
	if err := Run(); err != nil {
		log.Fatal("服务启动失败: ", err)
	}
}
