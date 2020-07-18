package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//bool connectionSucceed float32 time
func ping(ip string) (bool, float64) {
	startTime := time.Now()
	conn, err := net.DialTimeout("tcp", ip+":"+strconv.Itoa(defaultTcpPort), tcpConnectTimeout)
	if err != nil {
		return false, 0
	} else {
		var endTime = time.Since(startTime)
		var duration = float64(endTime.Microseconds()) / 1000.0
		_ = conn.Close()
		return true, duration
	}
}

//pingReceived pingTotalTime
func checkConnection(ip string) (int, float64) {
	pingRecv := 0
	var pingTime float64 = 0.0
	for i := 1; i <= failTime; i++ {
		pingSucceed, pingTimeCurrent := ping(ip)
		if pingSucceed {
			pingRecv++
			pingTime += pingTimeCurrent
		}
	}
	return pingRecv, pingTime
}

//return Success packetRecv averagePingTime specificIPAddr
func pingHandler(ip string, pingCount int) (bool, int, float64, string) {
	ipCanConnect := false
	pingRecv := 0
	var pingTime float64 = 0.0

	pingRecvCurrent, pingTimeCurrent := checkConnection(ip)
	if pingRecvCurrent != 0 {
		ipCanConnect = true
		pingRecv = pingRecvCurrent
		pingTime = pingTimeCurrent
	}

	if ipCanConnect {
		for i := failTime; i < pingCount; i++ {
			pingSuccess, pingTimeCurrent := ping(ip)
			if pingSuccess {
				pingRecv++
				pingTime += pingTimeCurrent
			}
		}
		return true, pingRecv, pingTime / float64(pingRecv), ip
	} else {
		return false, 0, 0, ""
	}
}

func pingGoroutine(wg *sync.WaitGroup, mutex *sync.Mutex, ip string, pingCount int, data *[]CloudflareIPData, pingRoutine chan bool) {
	defer func() {
		<-pingRoutine
		wg.Done()
	}()

	success, pingRecv, pingTimeAvg, currentIP := pingHandler(ip, pingCount)
	if success {
		mutex.Lock()
		var cfdata CloudflareIPData
		cfdata.ip = currentIP
		cfdata.pingReceived = pingRecv
		cfdata.pingTime = pingTimeAvg
		cfdata.pingCount = pingCount
		*data = append(*data, cfdata)
		mutex.Unlock()
	}
}

//bool : can download,float32 downloadSpeed
func DownloadHandler(ctx context.Context, url, ip string, downloadDataSize *int64) (bool, int64) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, 0
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(c context.Context, network, addr string) (net.Conn, error) {
				conn, err := (&net.Dialer{}).DialContext(c, network, ip+":443")
				return conn, err
			},
			DisableKeepAlives: true,
		},
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko")
	resp, err := client.Do(req)
	if err != nil {
		return false, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if err != nil && err == io.EOF {
				*downloadDataSize += int64(n)
				//fmt.Println("resp eof err :",err,n)
				break
			} else if err != nil {
				//fmt.Println("resp err :",err)
				return false, 0
			}
			*downloadDataSize += int64(n)
		}
		return true, 0
	} else {
		return false, 0
	}
}

func speedGoRoutine(wg *sync.WaitGroup, mutex *sync.Mutex, url, ip string, downSecond int, data *CloudflareIPData, downloadRoutine chan bool) {
	defer func() {
		<-downloadRoutine
		wg.Done()
	}()
	d := time.Now().Add(time.Duration(downSecond) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()
	var downloadDataSize int64
	//t := time.Now()
	//i := 0
Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		default:
			//i++
			//fmt.Println("第i次下载：",i)
			DownloadHandler(ctx, url, ip, &downloadDataSize)
		}
	}
	//fmt.Println("download time", time.Since(t),float64(downloadDataSize) / 1024 / 1024)
	speed := float64(downloadDataSize) / 1024 / 1024 / float64(downSecond)
	mutex.Lock()
	data.downloadSpeed = speed
	mutex.Unlock()
}
