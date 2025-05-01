package utils

import (
	"errors"
	"os"
	"path/filepath"
)

func WriteFileWithDir(filePath string, data []byte, perm os.FileMode) error {
	// 获取文件夹路径
	dir := filepath.Dir(filePath)
	// 递归创建文件夹（如果不存在）
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// 写入文件
	return os.WriteFile(filePath, data, perm)
}

func FileExists(filePath string) (bool, error) {
	// 使用 os.Stat 获取文件信息
	info, err := os.Stat(filePath)
	if err == nil {
		// 文件存在
		return !info.IsDir(), nil // 确保是文件而不是目录
	}
	if errors.Is(err, os.ErrNotExist) {
		// 文件不存在
		return false, nil
	}
	// 其他错误（如权限问题）
	return false, err
}
