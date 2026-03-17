package tests

import (
	"context"
	"testing"
)

func TestExample(t *testing.T) {
	// 使用 RunTestSuite 运行测试
	RunTestSuite(func(ctx context.Context, ts *TestingSuite) error {
		// 这里可以使用 ts.Collections 进行测试
		t.Log("Testing suite initialized successfully")
		return nil
	})
}
