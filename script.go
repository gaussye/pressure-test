package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

/*
使用方式：
1. 长连接，connShort = false
2. 短连接，connShort = true

协程数：wokers := 10
单协程循环请求数：count := 100000
*/

var (
	connShort = true
	closeFlag = "close"
)

func createClient(short bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 1 * time.Second,
			}).DialContext,
			DisableKeepAlives: short,
		},
		Timeout: 2 * time.Second,
	}
}

func requestServer(count int, ch chan<- string) {
	// 短连接测试
	client := createClient(connShort)
	for i := 0; i < count; i++ {
		resp, err := client.Get("http://[2406:da14:f5e:e001:e075::]")
		if err != nil {
			ch <- fmt.Sprint(err)
			continue
		}
		pod, _ := ioutil.ReadAll(resp.Body)
		ch <- string(pod)
		defer resp.Body.Close()
	}
	ch <- closeFlag
}

func main() {
	var worker, count int
	var shortFlag bool
	flag.IntVar(&worker, "worker", 10, "协程数")
	flag.IntVar(&count, "count", 10000, "每个协程请求数")
	flag.BoolVar(&shortFlag, "short", true, "使用短连接")
	flag.Parse()
	connShort = shortFlag
	ch := make(chan string)
	//开始计时
	begin := time.Now()
	fmt.Println("开始时间:", begin)
	for i := 0; i < worker; i++ {
		go requestServer(count, ch)
	}
	result := make(map[string]int)
	for {
		r := <-ch
		if _, ok := result[r]; !ok {
			result[r] = 0
		}
		result[r] += 1
		if v, ok := result[closeFlag]; ok {
			if v == worker {
				break
			}
		}
	}
	end := time.Now()
	fmt.Println("结束时间:", end, time.Since(begin))
	for k, v := range result {
		fmt.Printf("pod: %s, count: %d \n", k, v)
	}
}
