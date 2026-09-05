package reader

import (
	"bytes"
	"homelab-reader/pkg/models"
	"io"
	"os"
)

func ReadTXTChunk(filePath string, offset int64, length int64) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//规定文件的起点和切片大小
	sectionReader := io.NewSectionReader(file, offset, length)

	//开辟出一片length大小的空间
	buf := make([]byte, length)

	//将规定大小的切片内存注入到buf空间内
	n, err := sectionReader.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	//将buf的空间中的0~n之内的所有字节都切割进入validBuf里面，当出现io.EOF的时候，直接返还在这之前的字节部分
	validBuf := buf[:n]
	if err == io.EOF {
		return validBuf, nil
	}

	//确定\n的位置，并将其定位到合适的下标数字，最后输出0~lastLF,也就是\n的下标位置之前的所有字节
	lastLF := bytes.LastIndexByte(validBuf, '\n')
	if lastLF != -1 {
		return validBuf[:lastLF+1], nil
	}

	return validBuf, nil
}

func ValidBook(books []models.Book) bool {
	for _, book := range books {
		if book.Title == "" || book.FilePath == "" || book.Format == "" || book.Size == 0 {
			return false
		}
	}
	return true
}
