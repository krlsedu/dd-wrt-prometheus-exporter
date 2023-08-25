package main

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	activeWirelessClientsLabel = prometheus.NewDesc(
		prometheus.BuildFQName("wrt", "wireless", "active_clients_total"),
		"The total number of active wireless clients",
		[]string{"interface"},
		nil,
	)
	txPowerLabel = prometheus.NewDesc(
		prometheus.BuildFQName("wrt", "wireless", "tx_power"),
		"The current transmission power in mW",
		nil,
		nil,
	)
	wifiRateLabel = prometheus.NewDesc(
		prometheus.BuildFQName("wrt", "wireless", "wifi_rate"),
		"The current Wi-Fi rate",
		nil,
		nil,
	)
	packetsLabel = prometheus.NewDesc(
		prometheus.BuildFQName("wrt", "wireless", "packets"),
		"Wi-Fi packet RX/TX transmissions",
		[]string{"direction", "quality"},
		nil,
	)
	loadAvgLabel = prometheus.NewDesc(
		prometheus.BuildFQName("wrt", "", "load_average"),
		"Router load average",
		[]string{"avg"},
		nil,
	)
	memoryLabel = prometheus.NewDesc(
		prometheus.BuildFQName("wrt", "", "memory"),
		"Router load average",
		[]string{"type", "status"},
		nil,
	)
	interfaceBytesLabel = prometheus.NewDesc(
		prometheus.BuildFQName("wrt", "interface", "bytes_count"),
		"Interface bytes count",
		[]string{"interface", "direction"},
		nil,
	)
	interfaceRateBytesLabel = prometheus.NewDesc(
		prometheus.BuildFQName("wrt", "interface", "bytes_rate"),
		"Interface bytes rate",
		[]string{"interface", "direction"},
		nil,
	)
	interfaceIntervalTimeLabel = prometheus.NewDesc(
		prometheus.BuildFQName("wrt", "interface", "interval_time"),
		"Interface interval time",
		[]string{"interface"},
		nil,
	)
)

// WRTExporter is the exporter for DD-WRT router using web APIs
type WRTExporter struct {
	routerURL  string
	username   string
	password   string
	interfaces []string
}

// NewWRTExporter returns a new WRT exporter
func NewWRTExporter(endpoint string, username string, password string, interfaces []string) *WRTExporter {
	return &WRTExporter{
		routerURL:  endpoint,
		username:   username,
		password:   password,
		interfaces: interfaces,
	}
}

// Describe populates the prometheus label section
func (e *WRTExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- activeWirelessClientsLabel
	ch <- txPowerLabel
	ch <- wifiRateLabel
	ch <- packetsLabel
	ch <- loadAvgLabel
	ch <- memoryLabel
	ch <- interfaceBytesLabel
}

var timeIfAnt = make(map[string]int64)
var bytesIfAnt = make(map[string]int64)

// Collect retrieve all info from DD-WRT and returns the relevant metrics
func (e *WRTExporter) Collect(ch chan<- prometheus.Metric) {
	wlinfo, err := e.getStatusGeneral()
	if err != nil {
		log.Printf("Error calling general status API: %#v", err)
		return
	}

	var cntByInterface = make(map[string]int)
	for _, c := range *wlinfo.ActiveWirelessClients {
		if _, ok := cntByInterface[c.Interface]; !ok {
			cntByInterface[c.Interface] = 0
		}
		cntByInterface[c.Interface]++
	}

	for wlintf, num := range cntByInterface {
		ch <- prometheus.MustNewConstMetric(
			activeWirelessClientsLabel,
			prometheus.GaugeValue,
			float64(num),
			wlintf,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		txPowerLabel,
		prometheus.GaugeValue,
		float64(*wlinfo.RadioTX),
	)

	ch <- prometheus.MustNewConstMetric(
		wifiRateLabel,
		prometheus.GaugeValue,
		float64(*wlinfo.RadioRate),
	)

	// Packets stats
	ch <- prometheus.MustNewConstMetric(
		packetsLabel,
		prometheus.CounterValue,
		float64(wlinfo.NetPacketInfo.RXGoodPacket),
		"rx",
		"good",
	)

	ch <- prometheus.MustNewConstMetric(
		packetsLabel,
		prometheus.CounterValue,
		float64(wlinfo.NetPacketInfo.RXErrorPacket),
		"rx",
		"errors",
	)

	ch <- prometheus.MustNewConstMetric(
		packetsLabel,
		prometheus.CounterValue,
		float64(wlinfo.NetPacketInfo.TXGoodPacket),
		"tx",
		"good",
	)

	ch <- prometheus.MustNewConstMetric(
		packetsLabel,
		prometheus.CounterValue,
		float64(wlinfo.NetPacketInfo.TXErrorPacket),
		"tx",
		"errors",
	)

	// Load AVG
	ch <- prometheus.MustNewConstMetric(
		loadAvgLabel,
		prometheus.GaugeValue,
		float64(wlinfo.Uptime.LoadAvg1m),
		"1m",
	)
	ch <- prometheus.MustNewConstMetric(
		loadAvgLabel,
		prometheus.GaugeValue,
		float64(wlinfo.Uptime.LoadAvg5m),
		"5m",
	)
	ch <- prometheus.MustNewConstMetric(
		loadAvgLabel,
		prometheus.GaugeValue,
		float64(wlinfo.Uptime.LoadAvg15m),
		"15m",
	)

	// Memory
	ch <- prometheus.MustNewConstMetric(
		memoryLabel,
		prometheus.GaugeValue,
		float64(wlinfo.MemInfo.RAMTotal),
		"ram", "total",
	)
	ch <- prometheus.MustNewConstMetric(
		memoryLabel,
		prometheus.GaugeValue,
		float64(wlinfo.MemInfo.RAMUsed),
		"ram", "used",
	)
	ch <- prometheus.MustNewConstMetric(
		memoryLabel,
		prometheus.GaugeValue,
		float64(wlinfo.MemInfo.RAMFree),
		"ram", "free",
	)
	ch <- prometheus.MustNewConstMetric(
		memoryLabel,
		prometheus.GaugeValue,
		float64(wlinfo.MemInfo.RAMShared),
		"ram", "shared",
	)
	ch <- prometheus.MustNewConstMetric(
		memoryLabel,
		prometheus.GaugeValue,
		float64(wlinfo.MemInfo.RAMBuffers),
		"ram", "buffers",
	)
	ch <- prometheus.MustNewConstMetric(
		memoryLabel,
		prometheus.GaugeValue,
		float64(wlinfo.MemInfo.RAMCached),
		"ram", "cached",
	)

	ch <- prometheus.MustNewConstMetric(
		memoryLabel,
		prometheus.GaugeValue,
		float64(wlinfo.MemInfo.SWAPTotal),
		"swap", "total",
	)
	ch <- prometheus.MustNewConstMetric(
		memoryLabel,
		prometheus.GaugeValue,
		float64(wlinfo.MemInfo.SWAPUsed),
		"swap", "used",
	)
	ch <- prometheus.MustNewConstMetric(
		memoryLabel,
		prometheus.GaugeValue,
		float64(wlinfo.MemInfo.SWAPFree),
		"swap", "free",
	)

	for _, intf := range e.interfaces {
		intfstats, err := e.getStatusInterface(intf)
		if err != nil {
			log.Printf("Error calling interface %s status API: %#v", intf, err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			interfaceBytesLabel,
			prometheus.CounterValue,
			float64(intfstats.TXBytes),
			intf, "tx",
		)

		ch <- prometheus.MustNewConstMetric(
			interfaceBytesLabel,
			prometheus.CounterValue,
			float64(intfstats.RXBytes),
			intf, "rx",
		)

		unixNano := time.Now().UnixNano()
		if _, ok := bytesIfAnt[intf+"tx"]; !ok {
			bytesIfAnt[intf+"tx"] = intfstats.TXBytes
			timeIfAnt[intf+"tx"] = unixNano
		}
		if _, ok := bytesIfAnt[intf+"rx"]; !ok {
			bytesIfAnt[intf+"rx"] = intfstats.RXBytes
			timeIfAnt[intf+"tx"] = unixNano
		}

		var bytesRateTx = float64(intfstats.TXBytes - bytesIfAnt[intf+"tx"])
		var bytesRateRx = float64(intfstats.RXBytes - bytesIfAnt[intf+"rx"])
		if bytesRateTx < 0 || bytesRateRx < 0 {
			bytesRateTx = 0
			bytesRateRx = 0
		}

		intervalTime := unixNano - timeIfAnt[intf+"tx"]
		if intervalTime > 0 {
			bytesRateTx = (bytesRateTx / float64(intervalTime)) * 1000000000
			bytesRateRx = (bytesRateRx / float64(intervalTime)) * 1000000000
		}

		bytesIfAnt[intf+"tx"] = intfstats.TXBytes
		bytesIfAnt[intf+"rx"] = intfstats.RXBytes
		timeIfAnt[intf+"tx"] = unixNano
		timeIfAnt[intf+"rx"] = unixNano

		seconds := float64(intervalTime) / 1000000000

		ch <- prometheus.MustNewConstMetric(
			interfaceIntervalTimeLabel,
			prometheus.CounterValue,
			seconds,
			intf,
		)

		ch <- prometheus.MustNewConstMetric(
			interfaceRateBytesLabel,
			prometheus.GaugeValue,
			bytesRateTx,
			intf, "tx",
		)

		ch <- prometheus.MustNewConstMetric(
			interfaceRateBytesLabel,
			prometheus.GaugeValue,
			bytesRateRx,
			intf, "rx",
		)

		// TODO: add others?
	}
}
