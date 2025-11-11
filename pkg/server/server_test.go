// server/server_test.go
package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewRouter 测试 NewRouter 函数
func TestNewRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Error("NewRouter() = nil, want not nil")
	}
}

// TestHandleConversionInvalidMethod 测试 HandleConversion 处理无效方法
func TestHandleConversionInvalidMethod(t *testing.T) {
	// 创建测试服务器
	router := NewRouter()

	// 创建 GET 请求（应该被拒绝）
	req, _ := http.NewRequest("GET", "/api/v1/convert", nil)
	w := httptest.NewRecorder()

	// 发送请求
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleConversion() = %v, want %v", w.Code, http.StatusMethodNotAllowed)
	}

	// 检查响应体是否包含错误信息
	if !strings.Contains(w.Body.String(), "只允许 POST 方法") {
		t.Errorf("HandleConversion() = %v, want to contain '只允许 POST 方法'", w.Body.String())
	}
}

// TestHandleConversionEmptyBody 测试 HandleConversion 处理空请求体
func TestHandleConversionEmptyBody(t *testing.T) {
	// 创建测试服务器
	router := NewRouter()

	// 创建空请求体
	req, _ := http.NewRequest("POST", "/api/v1/convert", strings.NewReader(""))
	w := httptest.NewRecorder()

	// 发送请求
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleConversion() = %v, want %v", w.Code, http.StatusBadRequest)
	}

	// 检查响应体是否包含错误信息
	if !strings.Contains(w.Body.String(), "请求体为空") {
		t.Errorf("HandleConversion() = %v, want to contain '请求体为空'", w.Body.String())
	}
}

// TestStartWorkerPool 测试 StartWorkerPool 函数
func TestStartWorkerPool(t *testing.T) {
	// 保存原始的 JobQueue
	oldJobQueue := JobQueue

	// 恢复原始的 JobQueue
	defer func() {
		JobQueue = oldJobQueue
	}()

	// 调用 StartWorkerPool
	StartWorkerPool()

	// 验证 JobQueue 是否已创建
	if JobQueue == nil {
		t.Error("StartWorkerPool() failed to create JobQueue")
	}

	// 验证 JobQueue 的容量
	if cap(JobQueue) != MaxQueue {
		t.Errorf("JobQueue capacity = %v, want %v", cap(JobQueue), MaxQueue)
	}
}
