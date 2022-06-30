package file

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
)

func File() {
	fmt.Println("--------------")
}

func CopyFile(copyFileName, toFileName string, perm fs.FileMode) error {
	if copyFileName == "" || toFileName == "" {
		return fmt.Errorf("文件路径不能为空")
	}
	if !FileIsExist(copyFileName) {
		return fmt.Errorf("文件: %s ,不存在", copyFileName)
	}
	fileContent, err := ioutil.ReadFile(copyFileName)
	if err != nil {
		return fmt.Errorf("读取文件（%s）失败：%#v", copyFileName, err)
	}
	if err = ioutil.WriteFile(toFileName, fileContent, perm); err != nil {
		return fmt.Errorf("创建文件（%s）失败：%#v", toFileName, err)
	}
	return nil
}

func CopyFile2(srcFile, destFile string, perm fs.FileMode) error {
	if srcFile == "" || destFile == "" {
		return fmt.Errorf("文件路径不能为空")
	}
	if !FileIsExist(srcFile) {
		return fmt.Errorf("文件: %s ,不存在", srcFile)
	}
	srcF, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("文件（%s）读取失败: %v ", srcFile, err)
	}
	destF, err := os.OpenFile(destFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("文件（%s）读取失败: %v ", destFile, err)
	}
	defer srcF.Close()
	defer destF.Close()

	_, err = io.Copy(destF, srcF)
	if err != nil {
		return fmt.Errorf("文件拷贝失败: %v ", err)
	}
	return nil
}

func FolderIsExist(path string) bool {
	if s, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return s.IsDir()
	}
}

func FileIsExist(path string) bool {
	if s, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return !s.IsDir()
	}
}
