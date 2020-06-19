package eventprocessing

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

var (
	stdin *os.File
)

type MetricValue struct {
	Timestamp string   `json:"timestamp"`
	Name      string   `json:"name"`
	Entity    string   `json:"entity"`
	Value     float64  `json:"value"`
	Namespace string   `json:"namespace"`
	Tags      []string `json:"tags"`
}

type EventValue struct {
	Timestamp string           `json:"timestamp"`
	Entity    *corev2.Entity    `json:"entity"`
	Check     *corev2.Check     `json:"check"`
	Metrics   *corev2.Metrics   `json:"namespace"`
	Metadata  corev2.ObjectMeta `json:"metadata"`
}

// {
// 	"name": "avg_cpu",
// 	"value": "56.0",
// 	"timestamp": "2019-03-30 12:30:00.45",
// 	"entity": "demo_test_agent",
// 	"namespace": "demo_jk185160",
// 	"tags": [
//    "company_jkte001",
//    "site_1001"
// 	]
// }

func parseTimestamp(timestamp int64) (string, error) {
	stringTimestamp := strconv.FormatInt(timestamp, 10)
	if len(stringTimestamp) > 10 {
		stringTimestamp = stringTimestamp[:10]
	}
	t, err := strconv.ParseInt(stringTimestamp, 10, 64)
	if err != nil {
		return "", err
	}
	return time.Unix(t, 0).Format(time.RFC3339), nil
}

func buildTag(key string, value string, prefix string) string {
	if len(prefix) > 0 {
		return fmt.Sprintf("%s_%s_%s", prefix, key, value)
	}
	return fmt.Sprintf("%s_%s", key, value)
}

func GetMetricFromPoint(point *corev2.MetricPoint, entityID string, namespaceID string, entityLabels map[string]string, pointNameAsMetricName bool) (MetricValue, error) {
	var metric MetricValue

	metric.Entity = entityID
	metric.Namespace = namespaceID
	// Find metric name
	if pointNameAsMetricName {
		metric.Name = point.Name
	} else {
		nameField := strings.Split(point.Name, ".")
		metric.Name = nameField[0]
	}
	// Find metric timstamp
	unixTimestamp, err := parseTimestamp(point.Timestamp)
	if err != nil {
		return *new(MetricValue), fmt.Errorf("failed to validate event: %s", err.Error())
	}
	metric.Timestamp = unixTimestamp
	metric.Tags = make([]string, len(point.Tags)+len(entityLabels)+1)
	i := 0
	for _, tag := range point.Tags {
		metric.Tags[i] = buildTag(tag.Name, tag.Value, "")
		i++
	}
	for key, val := range entityLabels {
		metric.Tags[i] = buildTag(key, val, "entity")
		i++
	}
	metric.Tags[i] = fmt.Sprintf("sensu_entity_name_%s", entityID)
	metric.Value = point.Value
	return metric, nil
}

func ParseEventTimestamp(event *corev2.Event) (EventValue, error) {
	var eventValue EventValue

	eventValue.Entity = event.Entity
	eventValue.Check = event.Check
	eventValue.Metrics = event.Metrics
	eventValue.Metadata = event.ObjectMeta

	timestamp, err := parseTimestamp(event.Timestamp)
	if err != nil {
		return *new(EventValue), fmt.Errorf("failed to validate event: %s", err.Error())
	}

	eventValue.Timestamp = timestamp
	return eventValue, nil
}
