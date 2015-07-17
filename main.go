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

func addGauge(gauges []interface{}, row []string, prefix string, name string, index uint) (result []interface{}){
	if row[index] == "" {
		return gauges
	}

	name = strings.Join([]string{prefix, name}, ".")

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
		prefix := strings.Join([]string{"haproxy", row[0]}, ".")

		switch row[1] {
		case "BACKEND":
			prefix = strings.Join([]string{prefix, "backend"}, ".")
		case "FRONTEND":
			prefix = strings.Join([]string{prefix, "frontend"}, ".")
		}

		gauges := make([]interface{}, 0)
		gauges = addGauge(gauges, row, prefix, "qcur", 2)
		gauges = addGauge(gauges, row, prefix, "qmax", 3)
		gauges = addGauge(gauges, row, prefix, "scur", 4)
		gauges = addGauge(gauges, row, prefix, "smax", 5)
		gauges = addGauge(gauges, row, prefix, "downtime", 24)
		gauges = addGauge(gauges, row, prefix, "hrsp_1xx", 39)
		gauges = addGauge(gauges, row, prefix, "hrsp_2xx", 40)
		gauges = addGauge(gauges, row, prefix, "hrsp_3xx", 41)
		gauges = addGauge(gauges, row, prefix, "hrsp_4xx", 42)
		gauges = addGauge(gauges, row, prefix, "hrsp_5xx", 43)

		metrics := &librato.Metrics{Source: os.Getenv("LIBRATO_SOURCE"), Gauges: gauges}

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
