//====================================================================================================
// Copyright (C) 2016-present Anne Sakitin (Tianwan Ayana).                                          =
//                                                                                                   =
// Part of the NGA project.                                                                          =
// Licensed under the F2DLPR License.                                                                =
//                                                                                                   =
// YOU MAY NOT USE THIS FILE EXCEPT IN COMPLIANCE WITH THE LICENSE.                                  =
// Provided "AS IS", WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,                                   =
// unless required by applicable law or agreed to in writing.                                        =
//                                                                                                   =
// For details about the NGA project, visit: http://app.niggergo.work.                               =
// For details about the F2DLPR License terms and conditions, visit: http://license.fileto.download. =
//====================================================================================================

package nga

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func PathExist(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

func MoveFile(src, dst string) error {
	if dstDir := filepath.Dir(dst); !PathExist(dstDir) {
		if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
			return err
		}
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	atime, mtime := srcInfo.ModTime(), srcInfo.ModTime()
	defer func() {
		if PathExist(dst) {
			_ = os.Chtimes(dst, atime, mtime)
		}
	}()
	if err = os.Rename(src, dst); err == nil {
		return nil
	} else if err = CopyFile(src, dst); err != nil {
		return err
	} else if err = os.Remove(src); err != nil {
		return err
	}
	return nil
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func IsEmptyDir(path string) bool {
	dir, err := os.Open(path)
	if err != nil {
		return false
	}
	defer dir.Close()
	entries, err := dir.Readdirnames(0)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func IsEmptyFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Size() == 0
}

func IsHiddenPath(path string) bool {
	path = filepath.Clean(path)
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		_ = os.Remove(dst)
		return err
	}
	atime, mtime := srcInfo.ModTime(), srcInfo.ModTime()
	_ = os.Chtimes(dst, atime, mtime)
	return nil
}

func CopyDir(src, dst string) error {
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		info, err := dir.Info()
		if err != nil {
			return err
		}
		if dir.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return CopyFile(path, dstPath)
	})
}
