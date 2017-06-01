package main

// This package provides a utility to scrape the output of the Enphase Envoy
// Home Gateway and write it or stream it to various sinks.

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/html"

	"bytes"
	"strings"

	"github.com/golang/glog"
)

var ip string

func init() {
	flag.StringVar(&ip, "ip", "192.168.1.2", "IP of Envoy Gateway")
}

func main() {
	glog.Infof("Starting Envoy Scraper")
	flag.Parse()

	url := fmt.Sprintf("http://%s/production?locale=en", ip)

	for {
		data, err := getPage(url)
		if err != nil {
			glog.Warningf("Failed to get data (%s)", err)
			time.Sleep(10 * time.Second)
			continue
		}
		glog.V(4).Infof("Scanning %d byte document", len(data))
		tokenScan(data, dataCallback)
		time.Sleep(60 * time.Second)
	}
}

func dataCallback(s string) {
	glog.V(2).Infof(s)
	// TODO = stream this into a time series store like TSDB
}

func tokenScan(data []byte, callback func(string)) {
	reader := bytes.NewBuffer(data)
	scanner := html.NewTokenizer(reader)
	for {
		tt := scanner.Next()

		switch {

		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.TextToken:
			text := scanner.Token().Data
			text = strings.TrimSpace(text)
			if strings.HasSuffix(text, "kW") {
				callback(text)
			}
		}
	}
}

func getPage(url string) ([]byte, error) {
	var data []byte
	resp, err := http.Get(url)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
