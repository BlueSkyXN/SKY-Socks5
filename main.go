package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// loadURLs 函数用于从 urls.txt 文件中加载代理源 URL 列表
func loadURLs() ([]string, error) {
	// 检查 urls.txt 文件是否存在，如果不存在则创建一个空的文件
	if _, err := os.Stat("urls.txt"); os.IsNotExist(err) {
		file, err := os.Create("urls.txt")
		if err != nil {
			return nil, fmt.Errorf("创建urls.txt文件失败: %v", err)
		}
		file.Close()
		fmt.Println("已创建空的urls.txt文件，请添加代理源URL后再继续")
		os.Exit(1)
	}

	file, err := os.Open("urls.txt")
	if err != nil {
		return nil, fmt.Errorf("打开urls.txt文件失败: %v", err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}

	if len(urls) == 0 {
		fmt.Println("urls.txt文件为空，请添加代理源URL后再继续")
		os.Exit(1)
	}

	return urls, nil
}

// fetchAndSave 函数用于从代理源 URL 列表中抓取代理，并将去重后的代理保存到 unique_proxies.txt 文件中。
func fetchAndSave() {
	urls, err := loadURLs()
	if err != nil {
		fmt.Println(err)
		return
	}

	allProxies := []string{}
	uniqueProxies := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("抓取 %s 时出错: %v\n", url, err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("读取 %s 响应时出错: %v\n", url, err)
				return
			}

			proxies := strings.Split(string(body), "\n")
			mu.Lock()
			allProxies = append(allProxies, proxies...)

			for _, proxy := range proxies {
				if proxy != "" {
					uniqueProxies[proxy] = true
				}
			}
			mu.Unlock()

			fmt.Printf("从 %s 获取并添加代理成功\n", url)
		}(url)
	}

	wg.Wait()

	file, err := os.Create("unique_proxies.txt")
	if err != nil {
		fmt.Printf("创建文件时出错: %v\n", err)
		return
	}
	defer file.Close()

	for proxy := range uniqueProxies {
		file.WriteString(proxy + "\n")
	}

	fmt.Printf("一共收集了 %d 个代理\n", len(allProxies))
	fmt.Printf("去重后剩余 %d 个代理\n", len(uniqueProxies))
	fmt.Println("代理列表已保存到 unique_proxies.txt")
}

// checkProxy 函数使用代理访问 https://one.one.one.one，如果响应状态码为 200，则认为代理可用。
func checkProxy(proxy string) bool {
	proxyURL, err := url.Parse("socks5://" + proxy)
	if err != nil {
		return false
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://one.one.one.one")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// validateProxies 函数用于验证 unique_proxies.txt 文件中的代理，并将可用的代理保存到 validated_proxies.txt 文件中。
func validateProxies() {
	file, err := os.Open("unique_proxies.txt")
	if err != nil {
		fmt.Printf("打开文件时出错: %v\n", err)
		return
	}
	defer file.Close()

	var proxies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxies = append(proxies, scanner.Text())
	}

	totalProxies := len(proxies)
	validProxies := make([]string, 0)
	var wg sync.WaitGroup
	var mu sync.Mutex
	progress := make(chan int)

	go func() {
		for i := 0; i < totalProxies; {
			select {
			case <-progress:
				i++
				printProgress(i, totalProxies)
			}
		}
	}()

	for _, proxy := range proxies {
		wg.Add(1)
		go func(proxy string) {
			defer wg.Done()
			if checkProxy(proxy) {
				mu.Lock()
				validProxies = append(validProxies, proxy)
				mu.Unlock()
			}
			progress <- 1
		}(proxy)
	}
	wg.Wait()

	validFile, err := os.Create("validated_proxies.txt")
	if err != nil {
		fmt.Printf("创建文件时出错: %v\n", err)
		return
	}
	defer validFile.Close()

	for _, proxy := range validProxies {
		validFile.WriteString(proxy + "\n")
	}

	fmt.Printf("\n可用代理已保存到 validated_proxies.txt\n")
	fmt.Printf("验证过后共有 %d 个可用代理\n", len(validProxies))
}

// printProgress 函数用于打印验证进度条。
func printProgress(current, total int) {
	barLength := 40
	progress := float64(current) / float64(total)
	block := int(float64(barLength) * progress)
	bar := strings.Repeat("#", block) + strings.Repeat("-", barLength-block)
	fmt.Printf("\r验证进度: [%s] %d/%d (%.2f%%)", bar, current, total, progress*100)
}

// main 函数是程序的入口点，它依次调用 fetchAndSave 和 validateProxies 函数，并记录每个操作的执行时间。
func main() {
	startTime := time.Now()
	fmt.Println("开始抓取代理，并保存到 unique_proxies.txt")
	// 调用 fetchAndSave 函数抓取代理
	fetchAndSave()
	endTime := time.Now()
	fmt.Printf("操作完成，共花费了 %.2f 秒\n", endTime.Sub(startTime).Seconds())

	startTime = time.Now()
	fmt.Println("开始验证代理，并保存到 validated_proxies.txt")
	// 调用 validateProxies 函数验证代理
	validateProxies()
	endTime = time.Now()
	fmt.Printf("操作完成，共花费了 %.2f 秒\n", endTime.Sub(startTime).Seconds())
}
