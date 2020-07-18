package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	t := time.Now()
	initConfig()
	var wg sync.WaitGroup
	var mu sync.Mutex
	var data = make([]CloudflareIPData, 0)
	ips := loadIp()
	pingRoutine := make(chan bool, Conf.pingRoutine)
	for _, ip := range ips {
		wg.Add(1)
		pingRoutine <- false
		go pingGoroutine(&wg, &mu, ip, Conf.pingCount, &data, pingRoutine)
	}
	wg.Wait()

	data = filterIpData(data)
	speedTestIpCount := Conf.speedTestCount
	if speedTestIpCount > len(data) {
		speedTestIpCount = len(data)
	}
	speedTestRoutine := make(chan bool, Conf.downloadRoutine)
	for i := 0; i < speedTestIpCount; i++ {
		wg.Add(1)
		speedTestRoutine <- false
		go speedGoRoutine(&wg, &mu, Conf.downloadUrl, data[i].ip, Conf.downloadSecond, &data[i], speedTestRoutine)
	}
	wg.Wait()
	fmt.Println("测试延迟和速度用时:",time.Since(t))
	sortBySpeedAndModifyDns(data[:speedTestIpCount])
	//fmt.Println(data[:speedTestIpCount])
	//fmt.Println(len(data[:speedTestIpCount]))
	ExportTxt("./result.txt", data)
	fmt.Println("*****************************************************")
}
