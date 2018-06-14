package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type mobileInfo struct {
	rows    string
	mobile  string
	addr    string
	channel string
}

var chanGet = make(chan mobileInfo, 500)
var chanSet = make(chan int, 500)
var wg sync.WaitGroup

func main() {
	var mobile = mobileInfo{}
	xlsx, err := excelize.OpenFile("./mobile.xlsx")
	if err != nil {
		fmt.Println("excelize.OpenFile eror", err)
		return
	}
	rows := xlsx.GetRows("Sheet1")
	go func() {
		for rowNumber, row := range rows {
			for _, colcell := range row {
				fmt.Println("will check the mobile : ", colcell)
				mobile.mobile = colcell
				mobile.rows = strconv.Itoa(rowNumber + 1)
				wg.Add(1)
				chanSet <- 1
				go get(mobile)
				break
			}
		}
		wg.Wait()
		close(chanGet)
	}()
	for info := range chanGet {
		xlsx.SetCellValue("Sheet1", "B"+info.rows, info.addr)
		xlsx.SetCellValue("Sheet1", "C"+info.rows, info.channel)
	}
	now := time.Now().Format("20060102150405")
	err = xlsx.SaveAs("./" + now + ".xlsx")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("按任意键退出")
	fmt.Scanf("1")
}

func get(mobile mobileInfo) {

	//推出要做的事情
	defer func() {
		<-chanSet
		wg.Done()
	}()
	var mobileMap []map[string]string
	var mobileArray [1]string
	mobileArray[0] = mobile.mobile
	mobileByte, err := json.Marshal(mobileArray)
	if err != nil {
		fmt.Println(" json.Marshal error", err)
		return
	}
	req, err := http.NewRequest("POST", "http://120.78.76.139:1235/api/checkMobile/", bytes.NewBuffer(mobileByte))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println("http.NewRequest error :", err)
		return
	}
	defer req.Body.Close()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("client.Do error", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil.readAll error:", err)
		return
	}
	err = json.Unmarshal(body, &mobileMap)
	if err != nil {
		fmt.Println("json.Unmarshal error", err)
		return
	}
	mobile.addr = mobileMap[0]["addr"]
	mobile.channel = mobileMap[0]["checkStatus"]
	switch mobileMap[0]["operator"] {
	case "0":
		mobile.channel = "未知"
	case "1":
		mobile.channel = "移动"
	case "2":
		mobile.channel = "联通"
	case "3":
		mobile.channel = "电信"
	}
	chanGet <- mobile
}
