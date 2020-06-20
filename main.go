package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/nixwiz/sensu-go-elasticsearch/lib/pkg/eventprocessing"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	DatedIndex            bool
	FullEventLogging      bool
	PointNameAsMetricName bool
	Index                 string
	TrustedCAFile         string
	InsecureSkipVerify    bool
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-go-elasticsearch",
			Short:    "The Sensu Go handler for metric and event logging in elasticsearch\nRequired:  Set the ELASTICSEARCH_URL env var with an appropriate connection url (https://user:pass@hostname:port)",
			Keyspace: "sensu.io/plugins/elasticsearch/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "dated_index",
			Env:       "",
			Argument:  "dated_index",
			Shorthand: "d",
			Default:   false,
			Usage:     "Should the index have the current date postfixed? ie: metric_data-2019-06-27",
			Value:     &plugin.DatedIndex,
		},
		{
			Path:      "full_event_logging",
			Env:       "",
			Argument:  "full_event_logging",
			Shorthand: "f",
			Default:   false,
			Usage:     "send the full event body instead of isolating event metrics",
			Value:     &plugin.FullEventLogging,
		},
		{
			Path:      "point_name_as_metric_name",
			Env:       "",
			Argument:  "point_name_as_metric_name",
			Shorthand: "p",
			Default:   false,
			Usage:     "use the entire point name as the metric name",
			Value:     &plugin.PointNameAsMetricName,
		},
		{
			Path:      "index",
			Env:       "",
			Argument:  "index",
			Shorthand: "i",
			Default:   "",
			Usage:     "index to use",
			Value:     &plugin.Index,
		},
		{
			Path:      "insecure-skip-verify",
			Env:       "",
			Argument:  "insecure-skip-verify",
			Shorthand: "s",
			Default:   false,
			Usage:     "skip TLS certificate verification (not recommended!)",
			Value:     &plugin.InsecureSkipVerify,
		},
		{
			Path:      "trusted-ca-file",
			Env:       "",
			Argument:  "trusted-ca-file",
			Shorthand: "t",
			Default:   "",
			Usage:     "TLS CA certificate bundle in PEM format",
			Value:     &plugin.TrustedCAFile,
		},
	}
)

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(event *corev2.Event) error {
	if len(plugin.Index) == 0 {
		return fmt.Errorf("No index specified")
	}
	return nil
}

func generateIndex() string {
	if plugin.DatedIndex {
		dt := time.Now()
		return fmt.Sprintf("%s-%s", plugin.Index, dt.Format("2006.01.02"))
	}
	return plugin.Index
}

func executeHandler(event *corev2.Event) error {
	if plugin.FullEventLogging {
		eventValue, err := eventprocessing.ParseEventTimestamp(event)
		if err != nil {
			return fmt.Errorf("error processing sensu event into eventValue: %v", err)
		}
		msg, err := json.Marshal(eventValue)
		if err != nil {
			return fmt.Errorf("error serializing metric data to json payload: %v", err)
		}
		err = sendElasticSearchData(string(msg), plugin.Index)
		if err != nil {
			return fmt.Errorf("error sending metric data to elasticsearch: %v", err)
		}
		return nil
	}

	if !event.HasMetrics() {
		return fmt.Errorf("event does not contain metrics")
	}
	for _, point := range event.Metrics.Points {
		metric, err := eventprocessing.GetMetricFromPoint(point, event.Entity.Name, event.Entity.Namespace, event.Entity.Labels, plugin.PointNameAsMetricName)
		if err != nil {
			return fmt.Errorf("error processing sensu event MetricPoints into MetricValue: %v", err)
		}
		msg, err := json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("error serializing metric data to json payload: %v", err)
		}
		err = sendElasticSearchData(string(msg), plugin.Index)
		if err != nil {
			return fmt.Errorf("error sending metric data to elasticsearch: %v", err)
		}
	}
	return nil
}

func sendElasticSearchData(metricBody string, index string) error {
	var cfg elasticsearch.Config

	if len(plugin.TrustedCAFile) > 0 {
		cert, err := ioutil.ReadFile(plugin.TrustedCAFile)
		if err != nil {
			return err
		}
		cfg.CACert = cert
	} else if plugin.InsecureSkipVerify {
		cfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		}
	}
	es, _ := elasticsearch.NewClient(cfg)
	req := esapi.IndexRequest{
		Index:   generateIndex(),
		Body:    strings.NewReader(metricBody),
		Refresh: "true",
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), es)
	if err != nil {
		return fmt.Errorf("Error getting response: %s", err)
	}
	if res.IsError() {
		return fmt.Errorf("[%s] Error indexing document ID=%d", res.Status(), 0)
	}
	return nil
}
