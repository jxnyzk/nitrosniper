package files

import (
	"bufio"
	"errors"
	"io"
	"os"
	"sniper/logger"
	"strings"
)

func ReadLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	defer f.Close()

	r := bufio.NewReader(f)
	bytes, lines := []byte{}, []string{}

	for {
		line, isPrefix, err := r.ReadLine()
		if err != nil {
			break
		}

		bytes = append(bytes, line...)
		if !isPrefix {
			str := strings.TrimSpace(string(bytes))

			if len(str) > 0 {
				lines = append(lines, str)
				bytes = []byte{}
			}
		}
	}

	if len(bytes) > 0 {
		lines = append(lines, string(bytes))
	}

	return lines, nil
}

func CreateFileIfNotExists(filePath string) {
	var _, err = os.Stat(filePath)

	if os.IsNotExist(err) {
		var file, err = os.Create(filePath)
		if err != nil {
			return
		}
		defer file.Close()
	}
}

func AppendFile(filePath string, Content string) {
	File, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("Failed to append to file", logger.FieldString("path", filePath), logger.FieldAny("error", err))
		return
	}

	defer File.Close()

	_, err = File.WriteString(Content + "\n")
	if err != nil {
		logger.Error("Failed to append to file", logger.FieldString("path", filePath), logger.FieldAny("error", err))
		return
	}
}

func OverwriteFile(filePath string, Content string) {
	File, err := os.Create(filePath)
	if err != nil {
		logger.Error("Failed to overwrite file", logger.FieldString("path", filePath), logger.FieldAny("error", err))
		return
	}

	defer File.Close()

	_, err = File.WriteString(Content)
	if err != nil {
		logger.Error("Failed to append to file", logger.FieldString("path", filePath), logger.FieldAny("error", err))
		return
	}
}
