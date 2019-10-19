package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

func main() {

	var (
		file    = flag.String("file", "", "File that contains URLs to download")
		workers = flag.Int("workers", 1, "Amount of workers performing downloads, IE concurrent downloads")
		url     = flag.String("url", "", "Specifies a specific url to download")
	)
	flag.Parse()

	if *workers < 1 {
		panic(errors.New("invalid value for workers"))
	}

	if *url != "" {
		err := downloadFile(*url, "")
		if err != nil {
			fmt.Println("Error downloading URL: ", *url, "error: ", err)
		}
	}

	if *file != "" {
		data, err := ioutil.ReadFile(*file)
		if err != nil {
			panic(err)
		}
		lines := strings.Split(strings.Replace(string(data), "\r\n", "\n", -1), "\n")
		var wg sync.WaitGroup
		sem := make(chan struct{}, *workers)
		for _, line := range lines {
			details := strings.Split(line, ";")
			if len(details) != 2 {
				panic(errors.New("unexpected file format"))
			}
			source := details[0]
			destination := details[1]
			wg.Add(1)
			go func(s, d string, sem chan struct{}) {
				defer func() { <-sem }()
				defer wg.Done()
				sem <- struct{}{}
				fmt.Println("Downloading URL:", s, "to", d)
				err := downloadFile(s, d)
				if err != nil {
					fmt.Println("ERROR:", err)
					return
				}
				time.Sleep(2 * time.Second)
				fmt.Println("Downloaded ", s)
			}(source, destination, sem)
		}
		wg.Wait()
	}
}

// downloadFile downloads a file over HTTP and persist
// it to the filesystem.
func downloadFile(source, destination string) error {
	resp, err := http.Get(source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if destination == "" {
		destination = path.Base(resp.Request.URL.Path)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
