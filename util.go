package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultTcpPort = 443
const tcpConnectTimeout = time.Millisecond * 350
const failTime = 4

type CloudflareIPData struct {
	ip            string
	pingTime      float64
	pingCount     int
	pingReceived  int
	recvRate      float64
	downloadSpeed float64
}

func (cf *CloudflareIPData) getRecvRate() float64 {
	if cf.recvRate == 0 {
		cf.recvRate = float64(cf.pingReceived) / float64(cf.pingCount) * 100
	}
	return cf.recvRate
}

func convertExportData(data []CloudflareIPData) [][]string {
	result := make([][]string, 0)
	for _, v := range data {
		result = append(result, v.toString())
	}
	return result
}

func (cf *CloudflareIPData) toString() []string {
	result := make([]string, 6)
	result[0] = cf.ip
	result[1] = strconv.Itoa(cf.pingCount)
	result[2] = strconv.Itoa(cf.pingReceived)
	result[3] = strconv.FormatFloat(cf.getRecvRate(), 'f', 2, 64)
	result[4] = strconv.FormatFloat(cf.pingTime, 'f', 2, 64)
	result[5] = strconv.FormatFloat(cf.downloadSpeed, 'f', 2, 64)
	return result
}

func ExportTxt(filepath string, data []CloudflareIPData) {
	if len(data) > 5 {
		data = data[:5]
	}
	exportData := convertExportData(data)
	t := time.Now()
	str := fmt.Sprintln(exportData, "   ---   ", t.Format("2006-01-02 15:04:05"))
	if !Conf.isOutputTxt {
		fmt.Println(str)
		return
	}
	txt, _ := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend)
	defer txt.Close()
	n, err := txt.WriteString(str)
	if n != len(str) {
		panic(err)
	}
}

func filterIpData(data []CloudflareIPData) (res []CloudflareIPData) {
	sort.Slice(data, func(i, j int) bool {
		if data[i].getRecvRate() != data[j].getRecvRate() {
			return data[i].getRecvRate() > data[j].getRecvRate()
		}
		return data[i].pingTime < data[j].pingTime
	})
	for _, v := range data {
		if v.pingTime <= Conf.rttLimit && v.recvRate >= Conf.recvRateLimit {
			res = append(res, v)
		}
	}
	return
}

func loadIp() []string {
	buf, err := ioutil.ReadFile(Conf.ipFilename)
	if err != nil {
		fmt.Println("read ip file err", err)
		panic(err)
	}

	ips := strings.Split(string(buf), "\n")
	ips = ips[:len(ips)-1]
	ipList := make([]string, 0)

	count := Conf.selectCountEveryIp
	if count <= 0 || count > 255 {
		panic("每个ip段选择的ip数量,不能为0且小于等于255")
	}

	rand.Seed(time.Now().UnixNano())
	for _, v := range ips {
		ip := strings.Split(v, ".")
		for i := 0; i < count; i++ {
			num := rand.Intn(254) + 1
			ipList = append(ipList, fmt.Sprintf("%s.%s.%s.%v", ip[0], ip[1], ip[2], num))
		}
	}

	return ipList
}

func sortBySpeedAndModifyDns(data []CloudflareIPData) {
	if len(data) == 0 {
		return
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].downloadSpeed > data[j].downloadSpeed
	})
	if !DnsConf.modifyEnable {
		fmt.Println("不需要修改dns")
		return
	}

	ip := data[0].ip
	form := make(url.Values)
	form["login_token"] = []string{DnsConf.dnspodToken}
	form["domain"] = []string{DnsConf.domain}
	form["sub_domain"] = []string{DnsConf.subDomain}
	form["record_id"] = []string{DnsConf.recordId}
	form["record_type"] = []string{DnsConf.recordType}
	form["record_line"] = []string{DnsConf.recordLine}
	form["value"] = []string{ip}

	if data[0].downloadSpeed >= DnsConf.speedLimit {
		_, _ = http.PostForm("https://dnsapi.cn/Record.Modify", form)
		fmt.Println("修改dns成功")
	} else {
		fmt.Println("ip不符合要求,修改dns失败:",ip)
	}
}
