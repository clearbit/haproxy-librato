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

func parseField(data string) (value float64) {
	value, err := strconv.ParseFloat(data, 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return value
}

func addGauge(gauges []interface{}, source string, prefix string, name string, data string) (result []interface{}) {
	if data == "" {
		return gauges
	}

	name = strings.Join([]string{prefix, name}, ".")

	return append(gauges, librato.Metric{Source: source, Name: name, Value: parseField(data)})
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
		source := os.Getenv("LIBRATO_SOURCE")
		prefix := strings.Join([]string{"haproxy", row[0]}, ".")

		switch row[1] {
		case "BACKEND":
			prefix = strings.Join([]string{prefix, "backend"}, ".")
		case "FRONTEND":
			prefix = strings.Join([]string{prefix, "frontend"}, ".")
		default:
			source = strings.Join([]string{source, row[1]}, ".")
			prefix = strings.Join([]string{prefix, "upstream"}, ".")
		}

		gauges := make([]interface{}, 0)
		gauges = addGauge(gauges, source, prefix, "qcur", row[2])
		gauges = addGauge(gauges, source, prefix, "qmax", row[3])
		gauges = addGauge(gauges, source, prefix, "scur", row[4])
		gauges = addGauge(gauges, source, prefix, "smax", row[5])
		gauges = addGauge(gauges, source, prefix, "downtime", row[24])

		gauges = addGauge(gauges, source, prefix, "hrsp_1xx", row[39])
		gauges = addGauge(gauges, source, prefix, "hrsp_2xx", row[40])
		gauges = addGauge(gauges, source, prefix, "hrsp_3xx", row[41])
		gauges = addGauge(gauges, source, prefix, "hrsp_4xx", row[42])
		gauges = addGauge(gauges, source, prefix, "hrsp_5xx", row[43])

		metrics := &librato.Metrics{Gauges: gauges}

		err := client.PostMetrics(metrics)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	client := librato.Client{os.Getenv("LIBRATO_USER"), os.Getenv("LIBRATO_TOKEN")}
	ticker := time.Tick(30 * time.Second)
	for _ = range ticker {
		poll(client)
	}
}
