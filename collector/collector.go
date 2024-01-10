package collector

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hatamiarash7/netflow-exporter/config"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
	"github.com/tehmaze/netflow"
	"github.com/tehmaze/netflow/netflow5"
	"github.com/tehmaze/netflow/netflow9"
	"github.com/tehmaze/netflow/session"
)

var lastProcessed = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "netflow_last_processed",
		Help: "Unix timestamp of the last processed netflow metric.",
	},
)

type sample struct {
	Labels      map[string]string
	Counts      map[string]float64
	TimestampMs int64
}

// Collector is the main collector type
type Collector struct {
	Config  config.Config
	Channel chan *sample
	Samples map[string]*sample
	Mutex   *sync.Mutex
}

type timeConstMetric struct {
	Time   int64
	Metric prometheus.Metric
}

// NewCollector will define new NetFlow collector instance
func NewCollector(cfg config.Config) *Collector {
	c := &Collector{
		Config:  cfg,
		Channel: make(chan *sample, 0),
		Samples: map[string]*sample{},
		Mutex:   &sync.Mutex{},
	}
	go c.process()

	return c
}

// Reader will read NetFlow UDP packets from the socket
func (c *Collector) Reader(udpSock *net.UDPConn) {
	defer udpSock.Close()
	decoders := make(map[string]*netflow.Decoder)

	for {
		buf := make([]byte, 65535)
		chars, srcAddress, err := udpSock.ReadFromUDP(buf)

		if err != nil {
			log.Errorf("Error reading UDP packet from %s: %s", srcAddress, err)
			continue
		}

		timestampMs := int64(float64(time.Now().UnixNano()) / 1e6)

		d, found := decoders[srcAddress.String()]
		if !found {
			s := session.New()
			d = netflow.NewDecoder(s)
			decoders[srcAddress.String()] = d
		}

		m, err := d.Read(bytes.NewBuffer(buf[:chars]))
		if err != nil {
		}

		switch p := m.(type) {
		case *netflow5.Packet:
			for _, record := range p.Records {
				labels := prometheus.Labels{}
				counts := make(map[string]float64)

				labels["sourceIPv4Address"] = record.SrcAddr.String()
				labels["destinationIPv4Address"] = record.DstAddr.String()
				labels["sourceTransportPort"] = strconv.FormatUint(uint64(record.SrcPort), 10)
				labels["destinationTransportPort"] = strconv.FormatUint(uint64(record.DstPort), 10)
				counts["packetDeltaCount"] = float64(record.Packets)
				counts["octetDeltaCount"] = float64(record.Bytes)
				labels["protocolIdentifier"] = strconv.FormatUint(uint64(record.Protocol), 10)
				labels["tcpControlBits"] = strconv.FormatUint(uint64(record.TCPFlags), 10)
				labels["bgpSourceAsNumber"] = strconv.FormatUint(uint64(record.SrcAS), 10)
				labels["bgpDestinationAsNumber"] = strconv.FormatUint(uint64(record.DstAS), 10)
				labels["sourceIPv4PrefixLength"] = strconv.FormatUint(uint64(record.SrcMask), 10)
				labels["destinationIPv4PrefixLength"] = strconv.FormatUint(uint64(record.DstMask), 10)

				if (len(counts) > 0) && (len(labels) > 0) {
					labels["From"] = srcAddress.IP.String()
					labels["NetflowVersion"] = "5"

					sample := &sample{
						Labels:      labels,
						Counts:      counts,
						TimestampMs: timestampMs,
					}
					lastProcessed.Set(float64(time.Now().UnixNano()) / 1e9)
					c.Channel <- sample
				}
			}

		case *netflow9.Packet:
			for _, set := range p.DataFlowSets {
				for _, record := range set.Records {
					labels := prometheus.Labels{}
					counts := make(map[string]float64)

					for _, field := range record.Fields {
						if len(c.Config.Exclude) > 0 && regexp.MustCompile(c.Config.Exclude).MatchString(field.Translated.Name) {
							log.Debug(field, "is not using label")
						} else if regexp.MustCompile(c.Config.Include).MatchString(field.Translated.Name) {
							counts[field.Translated.Name] = float64(field.Translated.Value.(uint64))
							log.Debug(field, "is using metric")
						} else {
							labels[field.Translated.Name] = fmt.Sprintf("%v", field.Translated.Value)
						}

					}

					if (len(counts) > 0) && (len(labels) > 0) {
						labels["From"] = srcAddress.IP.String()
						labels["TemplateID"] = fmt.Sprintf("%d", record.TemplateID)
						labels["NetflowVersion"] = "9"
						sample := &sample{
							Labels:      labels,
							Counts:      counts,
							TimestampMs: timestampMs,
						}
						lastProcessed.Set(float64(time.Now().UnixNano()) / 1e9)
						c.Channel <- sample
					}
				}
			}
		default:
			log.Warn("packet is not supported version")
		}

	}
}

func makeName(l map[string]string) string {
	keys := make([]string, 0, len(l))
	for key := range l {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var name string
	for _, key := range keys {
		name += key + "=" + l[key]
	}

	return name
}

func (c *Collector) process() {
	ticker := time.NewTicker(time.Minute).C

	for {
		select {
		case sample := <-c.Channel:
			c.Mutex.Lock()

			_, ok := c.Samples[makeName(sample.Labels)]
			if !ok || (c.Samples[makeName(sample.Labels)].TimestampMs < sample.TimestampMs) {
				c.Samples[makeName(sample.Labels)] = sample
			}

			c.Mutex.Unlock()
		case <-ticker:
			ageLimit := int64(float64(time.Now().Add(-c.Config.SampleExpire).UnixNano()) / 1e6)
			c.Mutex.Lock()

			for k, sample := range c.Samples {
				if ageLimit >= sample.TimestampMs {
					delete(c.Samples, k)
				}
			}

			c.Mutex.Unlock()
		}
	}
}

// Describe will describe the metrics
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lastProcessed.Desc()
}

// Collect will collect the metrics
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ch <- lastProcessed
	c.Mutex.Lock()
	samples := make([]*sample, 0, len(c.Samples))

	for _, sample := range c.Samples {
		samples = append(samples, sample)
	}

	c.Mutex.Unlock()
	ageLimit := int64(float64(time.Now().Add(-c.Config.SampleExpire).UnixNano()) / 1e6)

	for _, sample := range samples {
		if ageLimit >= sample.TimestampMs {
			continue
		}

		for key, value := range sample.Counts {
			desc := ""

			if sample.Labels["TemplateID"] != "" {
				desc = fmt.Sprintf("netflow_%s_TemplateID%s_%s", sample.Labels["From"], sample.Labels["TemplateID"], key)
			} else {
				desc = fmt.Sprintf("netflow_%s_%s", sample.Labels["From"], key)
			}

			desc = strings.Replace(desc, ".", "", -1)
			log.Debug(desc)
			ch <- MustNewTimeConstMetric(
				prometheus.NewDesc(
					desc,
					fmt.Sprintf("netflow metric %s", key),
					[]string{}, sample.Labels,
				),
				prometheus.GaugeValue,
				value,
				sample.TimestampMs,
			)
		}
	}
}

// NewTimeConstMetric creates a new prometheus.Metric with a timestamp
func NewTimeConstMetric(desc *prometheus.Desc, valueType prometheus.ValueType,
	value float64, timestampMs int64) (prometheus.Metric, error) {
	return &timeConstMetric{
		Time:   timestampMs,
		Metric: prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, []string{}...),
	}, nil
}

// MustNewTimeConstMetric creates a new prometheus.Metric with a timestamp
func MustNewTimeConstMetric(desc *prometheus.Desc, valueType prometheus.ValueType,
	value float64, timestampMs int64) prometheus.Metric {
	m, err := NewTimeConstMetric(desc, valueType, value, timestampMs)

	if err != nil {
		panic(err)
	}

	return m
}

func (m *timeConstMetric) Desc() *prometheus.Desc {
	return m.Metric.Desc()
}

func (m *timeConstMetric) Write(out *dto.Metric) error {
	out.TimestampMs = &m.Time

	return m.Metric.Write(out)
}
