package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const intervalTime = 90 * 1000 * time.Millisecond // 暂停多久
const limitNumber = 38                            // 每请求多少个暂停一次
const requestURL = "https://github.com/signup_check/username"
const cookie = "has_recent_activity=1; _octo=GH1.1.1157388947.1533860098; logged_in=no; _ga=GA1.2.277350401.1533860107; _gat=1; tz=Asia%2FShanghai; _gh_sess=TU8wR0NwUWdtQ20rQS9TcElOY21NV1ZxTzBKaWxPcGZLM2dhalRrRWkvVllsNFdLN09NL2RpWDB0T1lOWmxGMldJZy9mYjhkSk9OWUNsV21xbkY0UWtNZWN0MElLZkliWEMwOHhFYkxLRWtXRGZvNk5jQjJVRGRFTlJZT3FhKzNyODR2ellKd1hYUmZad1ZGejVyM1JKQmZmenhsT1kySHBQcHFnUi9BWGFvbDYwaEN5cml0QXh0aEFWc2R1anlYUmxSVVNtU3lyZFIybnM0WmpwbE5QV29ZNFZPSysvUEJpV2QwbEpkMHl2TFhmenNTaEY4MmVVSHV6MnRLZldmajNJbFZGek1NZEFLUEcvREoxb1FONmJLWkJTcy9iVkRaQjNqT0Z1TlA1RnJhbUVwbTB4eFFsQnNlS1NMZzdlZ0c4U1VwWlp0clVlTy92dW11aXJQdU9uSkh0dk5HYjB0b3ZZdDNCbXBTRTVvPS0tdXdqQTE2bGhpTThYVHFUTi9IZ0gxQT09--55bcf49d8d2be868d41209b298906e6840b7cfcc"
const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3514.2 Safari/537.36"
const authenticityToken = "w0Wpi+cuA9ViGyZ2wpSnyrU/8GCoXo03B/ttYkqFr1lFvudkfTfZgDxkvjI7H3dgn3yvscGMQfW/WMGmCXq58w=="
const fileJSON = "./data.json"

// 储存集合(1296)
var assemble []string

// 过滤后的集合(936)(33696)
var assembleFilter []string

// 存放json数据
var dataJSON map[string]interface{}

func newfileUploadRequest(username string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	params := map[string]string{
		"authenticity_token": authenticityToken,
		"value":              username,
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err := writer.Close()

	req, err := http.NewRequest("POST", requestURL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	req.Header.Add("Accept", "text/html; fragment")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,vi;q=0.6")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", cookie)
	req.Header.Add("User-Agent", userAgent)

	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			dataJSON[username] = "ok"
			fmt.Printf("ok")
		} else {
			if len(body.String()) == 25 {
				dataJSON[username] = "already"
				fmt.Printf("already")
				// fmt.Printf("%v", body)
			} else if len(body.String()) < 100 {
				if strings.Contains(body.String(), "reserved") {
					dataJSON[username] = "reserved"
					fmt.Printf("reserved")
				} else {
					fmt.Printf("other")
				}
			} else {
				dataJSON[username] = "limit"
				fmt.Printf("limit")
			}
		}
		writeJSONFile()
	}
}

// 生成可能的集合
func calcAssemble() {
	data := [26]string{
		// "0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		"a", "b", "c", "d", "e", "f",
		"g", "h", "i", "j", "k", "l",
		"m", "n", "o", "p", "q", "r",
		"s", "t", "u", "v", "w", "x",
		"y", "z"}

	for _, v0 := range data {
		for _, v1 := range data {
			// assemble = append(assemble, v0+v1)
			for _, v2 := range data {
				assemble = append(assemble, v0+v1+v2)
			}
		}
	}
}

func filter(a []string) {
	for _, v := range a {
		// 判断首字是否为数字
		_, err := strconv.Atoi(v[0:1])
		// 不为数字则加入
		if err != nil {
			assembleFilter = append(assembleFilter, v)
		}
	}
}

// 递归发送请求
func cycleRequest(index int, limit int) {
	if index < len(assembleFilter) {
		if dataJSON[assembleFilter[index]] == nil ||
			dataJSON[assembleFilter[index]] == "limit" {

			newfileUploadRequest(assembleFilter[index])
		} else {
			fmt.Printf("cache")
			limit++
		}
		fmt.Printf("-")
		fmt.Printf(assembleFilter[index])
		fmt.Printf("|")

		index++
		if index == limit {
			time.Sleep(intervalTime)
			limit += limitNumber
		}
		cycleRequest(index, limit)
	} else {
		fmt.Println("扫描结束")
	}
}

// 读取json文件
func readJSONFile() []byte {
	fp, err := os.OpenFile(fileJSON, os.O_RDONLY, 0755)
	defer fp.Close()
	if err != nil {
		log.Fatal(err)
	}
	data := make([]byte, 100000)
	n, err := fp.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	return data[:n]
}

// 写入json文件
func writeJSONFile() {
	// return
	file, _ := os.OpenFile(fileJSON, os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()
	enc := json.NewEncoder(file)
	err := enc.Encode(dataJSON)
	if err != nil {
		log.Println("Error in encoding json")
	}
}

func initJSONFile() {
	_, err := os.Open(fileJSON)
	// 文件不存在则，创建
	if err != nil {
		outputFile, outputError := os.OpenFile(fileJSON, os.O_WRONLY|os.O_CREATE, 0666)
		if outputError != nil {
			fmt.Printf("An error occurred with file opening or creation\n")
			return
		}
		defer outputFile.Close()
		outputWriter := bufio.NewWriter(outputFile)
		outputString := "{}"

		outputWriter.WriteString(outputString)
		outputWriter.Flush()
	}
}

func main() {
	initJSONFile()

	data := readJSONFile()

	err := json.Unmarshal(data, &dataJSON)
	if err != nil {
		fmt.Println("json.Unmarshal err")
	}

	calcAssemble()
	filter(assemble)
	fmt.Printf("需请求的个数: %d\n", len(assembleFilter))
	cycleRequest(0, limitNumber)
}
