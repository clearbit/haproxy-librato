package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
)

func poll() {
	resp, err := http.Get(os.Getenv("HAPROXY_URL"))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)

	reader.Comment = '#'
	reader.FieldsPerRecord = -1

	data, err := reader.ReadAll()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client, err := statsd.NewClient("127.0.0.1:8125", "haproxy")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer client.Close()

	for _, each := range data {
		key := fmt.Sprintf("%s.%s.hrsp_2xx", each[0], each[1])
		value, err := strconv.ParseInt(each[40], 10, 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		client.Gauge(key, value, 1.0)
	}
}

func main() {
	for {
		poll()
		time.Sleep(5 * time.Second)
	}
}
