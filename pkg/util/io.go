package util

import (
	"encoding/csv"
	"log"
	"os"
)

// WriterCSV csv文件写入
func WriterCSV(path string, data [][]string) {

	//OpenFile读取文件，不存在时则创建，使用追加模式
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println("文件打开失败")
		panic(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	if err = writer.WriteAll(data); err != nil {
		log.Println("结果写入失败")
		panic(err)
	}

	writer.Flush() //刷新，不刷新是无法写入的
	log.Println("结果写入成功...")
}
