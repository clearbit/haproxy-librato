package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/samuel/go-librato/librato"
)

func field(row []string, index uint) (value float64) {
	value, err := strconv.ParseFloat(row[index], 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return value
}

func poll(client librato.Client) {
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

	for _, row := range data {
		if row[1] == "MONITOR" {
			continue
		}

		source := strings.Join([]string{row[0], os.Getenv("LIBRATO_SOURCE")}, "-")
		if row[1] != "BACKEND" {
			source = strings.Join([]string{source, row[1]}, "-")
		}

		gauges := make([]interface{}, 10)
		gauges[0] = librato.Metric{Name: "haproxy.qcur", Value: field(row, 2)}
		gauges[1] = librato.Metric{Name: "haproxy.qmax", Value: field(row, 3)}
		gauges[2] = librato.Metric{Name: "haproxy.scur", Value: field(row, 4)}
		gauges[3] = librato.Metric{Name: "haproxy.smax", Value: field(row, 5)}
		gauges[4] = librato.Metric{Name: "haproxy.downtime", Value: field(row, 24)}
		gauges[5] = librato.Metric{Name: "haproxy.hrsp_1xx", Value: field(row, 39)}
		gauges[6] = librato.Metric{Name: "haproxy.hrsp_2xx", Value: field(row, 40)}
		gauges[7] = librato.Metric{Name: "haproxy.hrsp_3xx", Value: field(row, 41)}
		gauges[8] = librato.Metric{Name: "haproxy.hrsp_4xx", Value: field(row, 42)}
		gauges[9] = librato.Metric{Name: "haproxy.hrsp_5xx", Value: field(row, 43)}

		metrics := &librato.Metrics{Source: source, Gauges: gauges}

		err := client.PostMetrics(metrics)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	client := librato.Client{os.Getenv("LIBRATO_USER"), os.Getenv("LIBRATO_TOKEN")}

	for {
		poll(client)
		time.Sleep(30 * time.Second)
	}
}
