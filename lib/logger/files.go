package logger

import (
	"fmt"
	"os"
)

// checkNotExist 检查文件或目录是否不存在
func checkNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

// checkPermission 检查是否有权限访问文件或目录
func checkPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

// isNotExistMkDir 如果目录不存在，则创建目录
func isNotExistMkDir(src string) error {
	if checkNotExist(src) {
		return mkDir(src)
	}
	return nil
}

// mkDir 创建目录，如果需要，会创建多级目录
func mkDir(src string) error {
	return os.MkdirAll(src, os.ModePerm) // os.ModePerm是0777，表示任何人都有读写执行权限
}

// mustOpen 打开或创建文件，如果目录不存在则创建目录，如果文件不存在则创建文件
func mustOpen(fileName, dir string) (*os.File, error) {
	if checkPermission(dir) {
		return nil, fmt.Errorf("permission denied dir: %s", dir)
	}

	if err := isNotExistMkDir(dir); err != nil {
		return nil, fmt.Errorf("error during make dir %s, err: %s", dir, err)
	}

	// os.O_APPEND表示以追加模式打开文件，os.O_CREATE表示如果文件不存在则创建，os.O_RDWR表示文件可读写
	// 0644表示文件所有者可读写，组用户和其他用户只能读取
	f, err := os.OpenFile(dir+string(os.PathSeparator)+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("fail to open file, err: %s", err)
	}

	return f, nil
}
