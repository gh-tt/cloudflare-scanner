package main

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	selectCountEveryIp int
	ipFilename         string
	pingRoutine        int
	pingCount          int
	speedTestCount     int
	downloadSecond     int
	downloadRoutine    int
	downloadUrl        string
	rttLimit           float64
	recvRateLimit      float64
	isOutputTxt        bool
}

type DnsConfig struct {
	modifyEnable bool
	dnspodToken,
	domain,
	subDomain,
	recordId,
	recordType,
	recordLine string
	speedLimit float64
}

var Conf *Config
var DnsConf *DnsConfig

func initConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("读取配置文件失败，请检查 config.yaml 配置文件是否存在: %v", err))
	}
	Conf = newConfig()
	DnsConf = newDnsConfig()
}
func newConfig() *Config {
	pingRoutine := viper.GetInt("pingRoutine")
	if pingRoutine <= 0 {
		pingRoutine = 100
	}
	pingCount := viper.GetInt("pingCount")
	if pingCount <= 0 {
		pingCount = 10
	}
	speedTestCount := viper.GetInt("downloadTestCount")
	if speedTestCount <= 0 {
		speedTestCount = 10
	}
	downloadSecond := viper.GetInt("downloadSecond")
	if downloadSecond <= 0 {
		downloadSecond = 10
	}
	downloadRoutine := viper.GetInt("downloadRoutine")
	if downloadRoutine <= 0 {
		downloadRoutine = 1
	}
	downloadUrl := viper.GetString("downloadUrl")
	if downloadUrl == "" {
		downloadUrl = "https://storage.idx0.workers.dev/Images/public-notion-06b4a73f-0d4e-4b8f-b273-77becf84a0b3.png"
	}
	rttLimit := viper.GetFloat64("rttLimit")
	if rttLimit <= 0 {
		rttLimit = 200
	}
	recvRateLimit := viper.GetFloat64("recvRateLimit")
	if recvRateLimit <= 0 || recvRateLimit > 100 {
		recvRateLimit = 0
	}
	ipFilename :=viper.GetString("ipFilename")
	if ipFilename=="" {
		ipFilename = "ip.txt"
	}

	return &Config{
		selectCountEveryIp: viper.GetInt("selectCountEveryIp"),
		ipFilename:         ipFilename,
		pingRoutine:        pingRoutine,
		pingCount:          pingCount,
		speedTestCount:     speedTestCount,
		downloadSecond:     downloadSecond,
		downloadRoutine:    downloadRoutine,
		downloadUrl:        downloadUrl,
		rttLimit:           rttLimit,
		recvRateLimit:      recvRateLimit,
		isOutputTxt:        viper.GetBool("isOutputTxt"),
	}
}
func newDnsConfig() *DnsConfig {
	speedLimit := viper.GetFloat64("dns.speedLimit")
	if speedLimit <= 0 {
		speedLimit = 0
	}
	return &DnsConfig{
		modifyEnable: viper.GetBool("dns.modifyEnable"),
		dnspodToken:  viper.GetString("dns.dnspodToken"),
		domain:       viper.GetString("dns.domain"),
		subDomain:    viper.GetString("dns.subDomain"),
		recordId:     viper.GetString("dns.recordId"),
		recordType:   viper.GetString("dns.recordType"),
		recordLine:   viper.GetString("dns.recordLine"),
		speedLimit:   speedLimit,
	}
}
