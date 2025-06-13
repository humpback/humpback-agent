package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetWorkDir 获取当前工作目录
func GetWorkDir() string {
	execPath, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("ExecutableDir: %s", err))
	}
	return filepath.Dir(execPath)
}

func Mkdir(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err1 := os.MkdirAll(dir, os.ModePerm); err1 != nil {
				return err1
			}
			return nil
		}
		return err
	}
	return nil
}

func FileExist(path string) bool {
	fpath, _ := filepath.Abs(path)
	_, err := os.Stat(fpath)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func EnsureDirAndForFile(filePath string) error {
	dir := filepath.Dir(filePath)
	info, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("get dir(%s) stat failed: %s", dir, err)
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir(%s) failed: %s", dir, err)
		}
		return nil
	}
	if !info.IsDir() {
		return fmt.Errorf("path(%s) not a dir", dir)
	}
	return nil
}
