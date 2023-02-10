package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

// Checks HTTP status of all paths
func GetPaths(host string, filename string, pathList []string, client *http.Client) []string {

	target := host + filename

	resp, err := client.Get(target)
	if err != nil {
		log.Fatalln(err)
	}

	resp.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:75.0) Gecko/20100101 Firefox/75.0")

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		sb := strings.ToLower(string(body))

		scanner := bufio.NewScanner(strings.NewReader(sb))

		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "path value") {
				index := strings.Index(scanner.Text(), "/")
				path := string(scanner.Text()[index:])
				index = strings.Index(path, "\"")
				path = string(path[:index])
				pathList = append(pathList, path)
			} else if strings.Contains(scanner.Text(), "filepath") {
				index := strings.Index(scanner.Text(), "/.")
				path := string(scanner.Text()[index:])
				index = strings.Index(path, "\"")
				path = string(path[:index])
				pathList = append(pathList, path)
			}
		}
	}

	return pathList
}

// Takes a slice of paths and returns a slice of paths that return 200 codes
func GetValidPaths(host string, pathList []string, threads int, client *http.Client) []string {
	var validPaths []string

	spin := spinner.New(spinner.CharSets[1], 100*time.Millisecond)
	spin.Prefix = "[*] Testing for valid file paths "
	spin.Start()

	sem := make(chan bool, threads)
	mut := &sync.Mutex{}

	for _, path := range pathList {
		sem <- true
		go func(path string) {
			resp, _ := client.Get(host + path)
			if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
				mut.Lock()
				validPaths = append(validPaths, path)
				mut.Unlock()
			}
			<-sem
		}(path)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	spin.Stop()

	fmt.Println("[!] Valid filepaths: ")

	return validPaths
}

// Download files from slice of succesful requests
func DownloadFiles(host string, pathList []string, client *http.Client) {
	u, _ := url.Parse(host)

	// Create base directory for target
	os.Mkdir(u.Host, os.ModePerm)

	for _, fullpath := range pathList {
		f, _ := url.Parse(fullpath)

		fileName := path.Base(fullpath)
		filePath := path.Dir(fullpath)

		// Take the full path of the file and mkdir for each leg of the path
		os.MkdirAll(u.Host+filePath, os.ModePerm)

		fullFilePath := u.Scheme + "://" + u.Host + f.Path

		// Create blank file
		file, err := os.Create(u.Host + filePath + "/" + fileName)
		if err != nil {
			log.Fatal(err)
		}
		// Put content on file
		resp, err := client.Get(fullFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		size, err := io.Copy(file, resp.Body)

		defer file.Close()

		fmt.Printf("Downloaded file %s with size %d\n", fileName, size)
	}
}

func CreateClient(proxy string) http.Client {
	if proxy == "NOPROXY" {
		tr := &http.Transport{
			MaxIdleConns:        30,
			MaxIdleConnsPerHost: 30,
			IdleConnTimeout:     30 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{Transport: tr}

		return *client

	} else {
		urlProxy, _ := url.Parse(proxy)

		tr := &http.Transport{
			MaxIdleConns:        30,
			MaxIdleConnsPerHost: 30,
			IdleConnTimeout:     30 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			Proxy:               http.ProxyURL(urlProxy),
		}

		client := &http.Client{Transport: tr}

		return *client
	}
}

func main() {
	hostPtr := flag.String("host", "REQUIRED", "Url to target. Example: https://example.com")
	proxyPtr := flag.String("proxy", "NOPROXY", "Proxy host and port. Example: http://127.0.0.1:8080")
	threadsPtr := flag.Int("threads", 50, "Number of concurrent threads to run. Example: 100")
	flag.Parse()

	if *hostPtr == "REQUIRED" {
		fmt.Println("A host is required. Example: -host https://example.com")
		os.Exit(0)
	}

	client := CreateClient(*proxyPtr)

	var pathList []string
	filenames := []string{"/.idea/workspace.xml", "/.idea/modules.xml", "/.idea/misc.xml"}

	// Take each file and return the slice of successes
	for _, filename := range filenames {
		pathList = GetPaths(*hostPtr, filename, pathList, &client)
	}

	fmt.Printf("[!] Found %d filepaths\n", len(pathList))

	validPaths := GetValidPaths(*hostPtr, pathList, *threadsPtr, &client)
	validPaths = append(validPaths, filenames...)

	DownloadFiles(*hostPtr, validPaths, &client)
}
