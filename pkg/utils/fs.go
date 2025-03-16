package utils

import (
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
