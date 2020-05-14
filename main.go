package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

const (
	namespace = "onlyoffice" // For Prometheus metrics.
)

var (
	listenAddress = flag.String("web.listen-address", ":9876", "Address on which to expose metrics.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	scrapeURI     = flag.String("scrape_uri", "http://localhost/info/info.json", "URI to the onlyoffice statistics info.")
	insecure      = flag.Bool("insecure", false, "Ignore onlyoffice server certificate if using https.")
)

type Exporter struct {
	URI    string
	mutex  sync.Mutex
	client *http.Client

	// OnlyOffice metrics.
	up                       *prometheus.Desc
	scrapeFailures           prometheus.Counter
	serverInfo               *prometheus.Desc
	editConnectionsLastHour  *prometheus.Desc
	viewConnectionsLastHour  *prometheus.Desc
	editConnectionsLastDay   *prometheus.Desc
	viewConnectionsLastDay   *prometheus.Desc
	editConnectionsLastWeek  *prometheus.Desc
	viewConnectionsLastWeek  *prometheus.Desc
	editConnectionsLastMonth *prometheus.Desc
	viewConnectionsLastMonth *prometheus.Desc
	licenseInfo              *prometheus.Desc
}

type OnlyofficeStats struct {
	Edit struct {
		Min uint `json:min`
		Avr uint `json:avr`
		Max uint `json:max`
	} `json:edit`
	View struct {
		Min uint `json:min`
		Avr uint `json:avr`
		Max uint `json:max`
	} `json:view`
}

type Onlyoffice struct {
	ConnectionsStat struct {
		Hour  OnlyofficeStats `json:hour`
		Day   OnlyofficeStats `json:day`
		Week  OnlyofficeStats `json:week`
		Month OnlyofficeStats `json:month`
	} `json:connectionsStat`
	LicenseInfo struct {
		Connections uint   `json:connections`
		HasLicense  bool   `json:hasLicense`
		BuildDate   string `json:buildDate`
		EndDate     string `json:endDate`
	} `json:licenseInfo`
	ServerInfo struct {
		BuildVersion string `json:buildVersion`
		BuildNumber  uint   `json:buildNumber`
	} `json:serverInfo`
}

// NewExporter allocates and initializes metrics
func NewExporter(uri string) *Exporter {
	return &Exporter{
		URI: uri,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the OnlyOffice server be reached",
			nil,
			nil),
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping onlyoffice.",
		}),
		editConnectionsLastHour: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "edit_connections_last_hour"),
			"Number of edit connections during last hour",
			[]string{"type"},
			nil,
		),
		viewConnectionsLastHour: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "view_connections_last_hour"),
			"Number of view connections during last hour",
			[]string{"type"},
			nil,
		),
		editConnectionsLastDay: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "edit_connections_last_day"),
			"Number of edit connections during last day",
			[]string{"type"},
			nil,
		),
		viewConnectionsLastDay: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "view_connections_last_day"),
			"Number of view connections during last day",
			[]string{"type"},
			nil,
		),
		editConnectionsLastWeek: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "edit_connections_last_week"),
			"Number of edit connections during last week",
			[]string{"type"},
			nil,
		),
		viewConnectionsLastWeek: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "view_connections_last_week"),
			"Number of view connections during last week",
			[]string{"type"},
			nil,
		),
		editConnectionsLastMonth: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "edit_connections_last_month"),
			"Number of edit connections during last month",
			[]string{"type"},
			nil,
		),
		viewConnectionsLastMonth: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "view_connections_last_month"),
			"Number of view connections during last month",
			[]string{"type"},
			nil,
		),
		licenseInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "license_info"),
			"License Information on OnlyOffice",
			[]string{"connections", "has_license", "build_date", "end_date"},
			nil,
		),
		serverInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_info"),
			"Server Information of OnlyOffice",
			[]string{"build_version", "build_number"},
			nil,
		),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
			},
		},
	}
}

// Helper.
func collectStat(ch chan<- prometheus.Metric, desc *prometheus.Desc, value float64, labelValue string) {
	ch <- prometheus.MustNewConstMetric(desc,
		prometheus.GaugeValue,
		value,
		labelValue)
}

// Request metrics to the onlyoffice server via http.
func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	req, err := http.NewRequest("GET", e.URI, nil)
	if err != nil {
		return fmt.Errorf("error building scraping request: %v", err)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return fmt.Errorf("error scraping onlyoffice: %v", err)
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		if err != nil {
			data = []byte(err.Error())
		}
		return fmt.Errorf("status %s (%d): %s", resp.Status, resp.StatusCode, data)
	}

	var onlyoffice Onlyoffice
	err = json.Unmarshal([]byte(data), &onlyoffice)
	if err != nil {
		return fmt.Errorf("not a valid json: %v", err)
		return err
	}

	collectStat(ch, e.editConnectionsLastHour, float64(onlyoffice.ConnectionsStat.Hour.Edit.Min), "min")
	collectStat(ch, e.editConnectionsLastHour, float64(onlyoffice.ConnectionsStat.Hour.Edit.Avr), "avr")
	collectStat(ch, e.editConnectionsLastHour, float64(onlyoffice.ConnectionsStat.Hour.Edit.Max), "max")
	collectStat(ch, e.viewConnectionsLastHour, float64(onlyoffice.ConnectionsStat.Hour.View.Min), "min")
	collectStat(ch, e.viewConnectionsLastHour, float64(onlyoffice.ConnectionsStat.Hour.View.Avr), "avr")
	collectStat(ch, e.viewConnectionsLastHour, float64(onlyoffice.ConnectionsStat.Hour.View.Max), "max")

	collectStat(ch, e.editConnectionsLastDay, float64(onlyoffice.ConnectionsStat.Day.Edit.Min), "min")
	collectStat(ch, e.editConnectionsLastDay, float64(onlyoffice.ConnectionsStat.Day.Edit.Avr), "avr")
	collectStat(ch, e.editConnectionsLastDay, float64(onlyoffice.ConnectionsStat.Day.Edit.Max), "max")
	collectStat(ch, e.viewConnectionsLastDay, float64(onlyoffice.ConnectionsStat.Day.View.Min), "min")
	collectStat(ch, e.viewConnectionsLastDay, float64(onlyoffice.ConnectionsStat.Day.View.Avr), "avr")
	collectStat(ch, e.viewConnectionsLastDay, float64(onlyoffice.ConnectionsStat.Day.View.Max), "max")

	collectStat(ch, e.editConnectionsLastWeek, float64(onlyoffice.ConnectionsStat.Week.Edit.Min), "min")
	collectStat(ch, e.editConnectionsLastWeek, float64(onlyoffice.ConnectionsStat.Week.Edit.Avr), "avr")
	collectStat(ch, e.editConnectionsLastWeek, float64(onlyoffice.ConnectionsStat.Week.Edit.Max), "max")
	collectStat(ch, e.viewConnectionsLastWeek, float64(onlyoffice.ConnectionsStat.Week.View.Min), "min")
	collectStat(ch, e.viewConnectionsLastWeek, float64(onlyoffice.ConnectionsStat.Week.View.Avr), "avr")
	collectStat(ch, e.viewConnectionsLastWeek, float64(onlyoffice.ConnectionsStat.Week.View.Max), "max")

	collectStat(ch, e.editConnectionsLastMonth, float64(onlyoffice.ConnectionsStat.Month.Edit.Min), "min")
	collectStat(ch, e.editConnectionsLastMonth, float64(onlyoffice.ConnectionsStat.Month.Edit.Avr), "avr")
	collectStat(ch, e.editConnectionsLastMonth, float64(onlyoffice.ConnectionsStat.Month.Edit.Max), "max")
	collectStat(ch, e.viewConnectionsLastMonth, float64(onlyoffice.ConnectionsStat.Month.View.Min), "min")
	collectStat(ch, e.viewConnectionsLastMonth, float64(onlyoffice.ConnectionsStat.Month.View.Avr), "avr")
	collectStat(ch, e.viewConnectionsLastMonth, float64(onlyoffice.ConnectionsStat.Month.View.Max), "max")

	ch <- prometheus.MustNewConstMetric(e.licenseInfo,
		prometheus.GaugeValue,
		1,
		fmt.Sprint(onlyoffice.LicenseInfo.Connections),
		fmt.Sprint(onlyoffice.LicenseInfo.HasLicense),
		onlyoffice.LicenseInfo.BuildDate,
		onlyoffice.LicenseInfo.EndDate)

	ch <- prometheus.MustNewConstMetric(e.serverInfo, prometheus.GaugeValue, 1,
		onlyoffice.ServerInfo.BuildVersion,
		fmt.Sprint(onlyoffice.ServerInfo.BuildNumber))

	return nil
}

// Collect fetches the statistics from the configured onlyoffice frontend, and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Errorf("Error scraping onlyoffice: %s", err)
		e.scrapeFailures.Inc()
		e.scrapeFailures.Collect(ch)
	}
	return
}

// Describe implements Collector.
// https://github.com/prometheus/client_golang/issues/140
// NOTE: I must confess that it is still not crystal clear in my mind! :)
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.serverInfo
	ch <- e.editConnectionsLastHour
	ch <- e.editConnectionsLastHour
	ch <- e.editConnectionsLastDay
	ch <- e.editConnectionsLastDay
	ch <- e.editConnectionsLastWeek
	ch <- e.editConnectionsLastWeek
	ch <- e.editConnectionsLastMonth
	ch <- e.editConnectionsLastMonth
	ch <- e.licenseInfo

	e.scrapeFailures.Describe(ch)
}

func main() {
	flag.Parse()

	exporter := NewExporter(*scrapeURI)
	prometheus.MustRegister(exporter)

	log.Infoln("Starting prometheus-onlyoffice-exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())
	log.Infof("Starting Server: %s", *listenAddress)
	log.Infof("Collect from: %s", *scrapeURI)

	http.Handle(*metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
