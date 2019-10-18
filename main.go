package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

func main() {

	var (
		file = flag.String("file", "downloads.txt", "File that contans URLs to download")
		//workers = flag.Int("workers", 0, "Amount of workers performing downloads")
	)
	flag.Parse()

	if *file != "" {
		data, err := ioutil.ReadFile(*file)
		if err != nil {
			panic(err)
		}
		lines := strings.Split(string(data), "\r\n")
		var wg sync.WaitGroup
		for _, line := range lines {
			details := strings.Split(line, ";")
			if len(details) != 2 {
				panic(errors.New("Unexpected file format"))
			}
			source := details[0]
			destination := details[1]
			wg.Add(1)
			go func(s, d string) {
				fmt.Println("Downloading URL:", s, "to", d)
				defer wg.Done()
				err := downloadFile(s, d)
				if err != nil {
					fmt.Println("ERROR:", err)
					return
				}
				fmt.Println("Downloaded ", s)
			}(source, destination)
		}
		wg.Wait()
	}
}

func downloadFile(source, destination string) error {
	resp, err := http.Get(source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
