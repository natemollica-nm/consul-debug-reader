package metrics

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ryanuber/columnize"
	"net/http"
	"regexp"
	"strings"
)

// TODO: Make the URL Agent Version adaptable (i.e., alter URL string to corresponding version)
const (
	telemetryURL = "https://developer.hashicorp.com/consul/docs/agent/telemetry"
)

type AgentTelemetryMetric struct {
	Name string
	Unit string
	Type string
}

func GetTelemetryMetrics() (string, []AgentTelemetryMetric, error) {
	// Define a data structure to store metric endpoints.
	telemetryMetrics := []string{"Metric\x1fUnit\x1fType"}

	// Send an HTTP GET request to the Consul telemetry metrics reference page.
	response, err := http.Get(telemetryURL)
	if err != nil {
		return "", []AgentTelemetryMetric{}, err
	}
	defer response.Body.Close()

	// Parse the HTML content.
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", []AgentTelemetryMetric{}, err
	}

	// Extract metric endpoints from the HTML.
	doc.Find("table tbody tr").Each(func(index int, rowHtml *goquery.Selection) {
		// Parse and extract metric endpoint.
		endpoint := rowHtml.Find("td:nth-child(1)").Text()
		metricUnit := rowHtml.Find("td:nth-child(3)").Text()
		metricType := rowHtml.Find("td:nth-child(4)").Text()
		if strings.HasPrefix(endpoint, "consul") {
			telemetryMetrics = append(telemetryMetrics, fmt.Sprintf("%s\x1f%s\x1f%s\x1f",
				endpoint, metricUnit, metricType))
		}
	})
	var telemetryInfo []AgentTelemetryMetric
	for i, line := range telemetryMetrics {
		infoSections := strings.Split(line, string([]byte{0x1f}))
		if len(infoSections) < 3 || i == 0 {
			continue
		}
		info := AgentTelemetryMetric{
			Name: infoSections[0],
			Unit: infoSections[1],
			Type: infoSections[2],
		}
		telemetryInfo = append(telemetryInfo, info)
	}
	// Build output string in columnized format for readability
	output, err := columnize.Format(telemetryMetrics, &columnize.Config{Delim: string([]byte{0x1f}), Glue: " "})
	if err != nil {
		return "", []AgentTelemetryMetric{}, nil
	}
	return output, telemetryInfo, nil
}

func ListMetrics() error {
	var latestMetrics string
	var err error
	if latestMetrics, _, err = GetTelemetryMetrics(); err != nil {
		return err
	}
	fmt.Printf("\nConsul Telemetry Metric Names (pulled from: %s)\n\n", telemetryURL)
	fmt.Println(latestMetrics)
	return nil
}

func ValidateMetricName(name string) error {
	var latestMetrics string
	var err error
	if latestMetrics, _, err = GetTelemetryMetrics(); err != nil {
		return err
	}
	// This metric name is dynamic and can be anything that the customer uses for service names
	reg := regexp.MustCompile(`^consul\.proxy\..+$`)
	if reg.MatchString(name) {
		fmt.Printf("built-in mesh proxy prefix used: %s\n", name)
		return nil
	}
	// list of metrics contains the name somewhere, return with no error
	if strings.Contains(latestMetrics, name) {
		return nil
	}
	return fmt.Errorf(fmt.Sprintf("[metrics-name-validation] '%s' not a valid telemetry metric name\n  visit: %s for full list of consul telemetry metrics", name, telemetryURL))
}
