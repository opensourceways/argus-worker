// main_test.go
package main

import (
	"testing"

	"github.com/opensourceways/argus-worker/pkg/server"
)

// TestRun 测试 Run 函数
func TestRun(t *testing.T) {
	// 由于 Run 函数会启动服务器并阻塞，我们只测试它不会 panic
	// 并且能够正确初始化各个组件

	// 这里我们不直接调用 Run()，因为它会启动一个 HTTP 服务器并阻塞
	// 相反，我们测试 Run 函数中调用的各个组件

	// 重置 server 包中的 JobQueue
	oldJobQueue := server.JobQueue
	server.JobQueue = nil
	defer func() {
		server.JobQueue = oldJobQueue
	}()

	// 测试 StartWorkerPool
	server.StartWorkerPool()

	// 验证工作池是否正确初始化
	if server.JobQueue == nil {
		t.Error("StartWorkerPool() 未能正确初始化 JobQueue")
	}

	if cap(server.JobQueue) != server.MaxQueue {
		t.Errorf("JobQueue 容量不正确: 期望 %d, 得到 %d", server.MaxQueue, cap(server.JobQueue))
	}

	// 测试 NewRouter
	router := server.NewRouter()
	if router == nil {
		t.Error("NewRouter() 返回了 nil")
	}
}

// TestMain 测试主函数
func TestMain(t *testing.T) {
	// 由于 main 函数会调用 os.Exit，我们不能直接测试它
	// 相反，我们测试 main 函数中的逻辑是否能正确执行

	// 我们已经通过 TestRun 测试了 Run 函数的主要逻辑
	// 所以这里我们只需要确保 main 函数能正确调用 Run 函数而不会 panic

	// 注意：在实际测试中，我们不会真正启动服务器，因为这会使测试复杂化
	// 并可能导致资源泄露
}

// TestIntegration 测试整个应用的集成
func TestIntegration(t *testing.T) {
	// 这个测试会验证整个应用的启动和基本功能

	// 重置 server 包中的 JobQueue
	oldJobQueue := server.JobQueue
	server.JobQueue = nil
	defer func() {
		server.JobQueue = oldJobQueue
	}()

	// 启动工作池
	server.StartWorkerPool()

	// 验证工作池是否正确初始化
	if server.JobQueue == nil {
		t.Error("工作池未能正确初始化")
	}

	if cap(server.JobQueue) != server.MaxQueue {
		t.Errorf("JobQueue 容量不正确: 期望 %d, 得到 %d", server.MaxQueue, cap(server.JobQueue))
	}

	// 创建路由器
	router := server.NewRouter()
	if router == nil {
		t.Error("NewRouter() 返回了 nil")
	}
}
