package utils

import (
	"archive/zip"
	// "compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func DeCompress(dst, src, filename string) error {
	if dst != "" { // 如果解压后不是放在当前目录就按照保存目录去创建目录
		if err := os.MkdirAll(dst, os.ModePerm); err != nil {
			Msg(err.Error())
			return err
		}
	}

	if strings.HasSuffix(src, ".gz") {
		return UnGzip(dst, src, filename)
	} else if strings.HasSuffix(src, ".zip") {
		return UnZip(dst, src)
	}
	return errors.New("不支持压缩文件类型")
}

// https://studygolang.com/articles/7481 https://www.cnblogs.com/smiler/p/7000200.html
func UnGzip(dst, src, filename string) error {
	_, e := Exec("/bin/sh", "-c", "gzip -d -c "+src+" > "+dst+"/"+filename)
	return e
}

// https://learnku.com/articles/23434/golang-learning-notes-five-archivezip-to-achieve-compression-and-decompression
func UnZip(dst, src string) (err error) {
	// 打开压缩文件，这个 zip 包有个方便的 ReadCloser 类型
	// 这个里面有个方便的 OpenReader 函数，可以比 tar 的时候省去一个打开文件的步骤
	zr, err := zip.OpenReader(src)
	defer zr.Close()
	if err != nil {
		return
	}

	for _, file := range zr.File { // 遍历 zr ，将文件写入到磁盘
		path := filepath.Join(dst, file.Name)

		if file.FileInfo().IsDir() { // 如果是目录，就创建目录
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return err
			}
			continue // 因为是目录，跳过当前循环，因为后面都是文件的处理
		}

		fr, err := file.Open() // 获取到 Reader
		if err != nil {        // 因为是在循环中，无法使用 defer ，直接放在最后
			fr.Close()
			return err
		}

		// 创建要写出的文件对应的 Write
		fw, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			fw.Close()
			fr.Close()
			return err
		}

		_, err = io.Copy(fw, fr)
		if err != nil {
			fw.Close()
			fr.Close()
			return err
		}

		// 将解压的结果输出
		// Msg("成功解压 %s ，共写入了 %d 个字符的数据\n", path, n)

		fw.Close()
		fr.Close()
	}
	return nil
}
