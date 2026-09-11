package reader

import (
	"homelab-reader/pkg/models"
	"os"
)

func ReadTXTChunk(filePath string, offset int64, length int64) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if length <= 0 {
		length = 1
	}

	// 向前多读 3 字节，以便把 offset 处被切断的半个 UTF-8 字符补完整
	readFrom := offset - 3
	if readFrom < 0 {
		readFrom = 0
	}
	buf := make([]byte, length+3)
	n, _ := file.ReadAt(buf, readFrom)
	data := buf[:n]
	if len(data) == 0 {
		return data, nil
	}

	// 起点对齐：跳过前导的 UTF-8 续字节（0x80~0xBF），保证从完整字符开头返回
	lead := 0
	for lead < len(data) && (data[lead]&0xC0) == 0x80 {
		lead++
	}
	data = data[lead:]

	// 终点对齐：找到最后一个完整字符的末尾下标
	// 注意：完整字符的末尾也可能是续字节，需先定位 rune 起始再判断是否完整
	s := len(data) - 1
	for s >= 0 && (data[s]&0xC0) == 0x80 {
		s--
	}
	if s < 0 {
		return nil, nil // 全为续字节，无有效字符
	}
	b := data[s]
	var need int
	switch {
	case b&0xF8 == 0xF0:
		need = 4
	case b&0xF0 == 0xE0:
		need = 3
	case b&0xE0 == 0xC0:
		need = 2
	default:
		need = 1 // ASCII
	}
	if s+need > len(data) {
		return data[:s], nil // 该字符不完整，去掉它的残留字节
	}
	return data[:s+need], nil
}

func ValidBook(books []models.Book) bool {
	for _, book := range books {
		if book.Title == "" || book.FilePath == "" || book.Format == "" || book.Size == 0 {
			return false
		}
	}
	return true
}
