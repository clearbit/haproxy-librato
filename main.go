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

func parseField(row []string, index uint) (value float64) {
	value, err := strconv.ParseFloat(row[index], 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return value
}

func addGauge(gauges []interface{}, row []string, name string, index uint) (result []interface{}){
	if row[index] == "" {
		return gauges
	}

	return append(gauges, librato.Metric{Name: name, Value: parseField(row, index)})
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
		source := strings.Join([]string{row[0], os.Getenv("LIBRATO_SOURCE")}, "-")
		if row[1] != "BACKEND" {
			source = strings.Join([]string{source, row[1]}, "-")
		}

		gauges := make([]interface{}, 0)
		gauges = addGauge(gauges, row, "haproxy.qcur", 2)
		gauges = addGauge(gauges, row, "haproxy.qmax", 3)
		gauges = addGauge(gauges, row, "haproxy.scur", 4)
		gauges = addGauge(gauges, row, "haproxy.smax", 5)
		gauges = addGauge(gauges, row, "haproxy.downtime", 24)
		gauges = addGauge(gauges, row, "haproxy.hrsp_1xx", 39)
		gauges = addGauge(gauges, row, "haproxy.hrsp_2xx", 40)
		gauges = addGauge(gauges, row, "haproxy.hrsp_3xx", 41)
		gauges = addGauge(gauges, row, "haproxy.hrsp_4xx", 42)
		gauges = addGauge(gauges, row, "haproxy.hrsp_5xx", 43)

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
