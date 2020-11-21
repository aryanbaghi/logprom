package logprom

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/aryanbaghi/logprom/internal/config"
)

type Container struct {
	ContainerID *regexp.Regexp
	Name        *regexp.Regexp
	Image       *regexp.Regexp
}

type Logprom struct {
	Container         Container
	LogFormat         *regexp.Regexp
	PrometheusMapping map[string]interface{}
}

type Track struct {
	ContainerID string
	Remove      bool
}

var tracking []Logprom
var watchingContainers map[string]bool
var metrics []interface{}

func TrackConfig(cpath string, registerer *prometheus.Registerer) {
	watchingContainers = map[string]bool{}
	if registerer == nil {
		registerer = &prometheus.DefaultRegisterer
	}
	c := config.Config{}
	err := config.LoadConfig(cpath, &c)
	if err != nil {
		panic(err)
	}
	parseConfig(&c, registerer)
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	track := make(chan Track)
	go func() {
		ctx := context.Background()
		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		for {
			select {
			case <-ticker.C:
				containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
				if err != nil {
					panic(err)
				}
				watchContainers(containers, track)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	go func() {
		for {
			if t, ok := <-track; ok {
				if t.Remove {
					delete(watchingContainers, t.ContainerID)
				} else {
					watchingContainers[t.ContainerID] = true
				}
			}
		}
	}()
}

func watchContainers(containers []types.Container, track chan Track) {
	for _, tc := range tracking {
		for _, cont := range containers {
			if _, ok := watchingContainers[cont.ID]; ok {
				continue
			}
			match := false
			if tc.Container.ContainerID != nil {
				if tc.Container.ContainerID.MatchString(cont.ID) {
					match = true
				}
			}
			if tc.Container.Name != nil {
				for _, name := range cont.Names {
					if tc.Container.Name.MatchString(name) {
						match = true
						break
					}
				}
			}
			if tc.Container.Image != nil {
				if tc.Container.Image.MatchString(cont.Image) {
					match = true
				}
			}
			if match {
				go trackContainer(cont.ID, cont.Names[0], tc.LogFormat, tc.PrometheusMapping, track)
			}
		}
	}
}

func parseConfig(c *config.Config, registerer *prometheus.Registerer) {
	tracking = make([]Logprom, len(c.Logprom))
	for i, l := range c.Logprom {
		tracking[i] = Logprom{
			Container:         Container{},
			LogFormat:         regexp.MustCompile(l.LogFormat),
			PrometheusMapping: map[string]interface{}{},
		}

		labels := []string{}
		if l.Container.ContainerID != "" || l.Container.Name != "" || l.Container.Image != "" {
			labels = append(labels, "container_name")
		}

		for i, label := range tracking[i].LogFormat.SubexpNames() {
			if i != 0 && label != "" {
				if strings.HasPrefix(label, "label_") {
					labels = append(labels, strings.Replace(label, "label_", "", 1))
				}
			}
		}

		for k, t := range l.PrometheusMapping {
			switch t {
			case config.Counter:
				m := prometheus.NewCounterVec(
					prometheus.CounterOpts{
						Name: fmt.Sprintf("logprom_%s", k),
					},
					labels,
				)
				(*registerer).MustRegister(m)
				tracking[i].PrometheusMapping[k] = m
			case config.Gauge:
				m := prometheus.NewGaugeVec(
					prometheus.GaugeOpts{
						Name: fmt.Sprintf("logprom_%s", k),
					},
					labels,
				)
				(*registerer).MustRegister(m)
				tracking[i].PrometheusMapping[k] = m
			case config.Histogram:
				m := prometheus.NewHistogramVec(prometheus.HistogramOpts{
					Name: fmt.Sprintf("logprom_%s", k),
					// Buckets: prometheus.LinearBuckets(20, 5, 5),
				},
					labels,
				)
				(*registerer).MustRegister(m)
				tracking[i].PrometheusMapping[k] = m

			}
		}
		if l.Container.ContainerID != "" {
			tracking[i].Container.ContainerID = regexp.MustCompile(l.Container.ContainerID)
		}
		if l.Container.Name != "" {
			tracking[i].Container.Name = regexp.MustCompile(l.Container.Name)
		}
		if l.Container.Image != "" {
			tracking[i].Container.Image = regexp.MustCompile(l.Container.Image)
		}
	}
}

func trackContainer(containerID, containerName string, logRegex *regexp.Regexp, metric map[string]interface{}, track chan Track) {
	log.Infof("start tracking container %s with id %s", containerName, containerID)
	track <- Track{ContainerID: containerID, Remove: false}
	defer func() {
		log.Infof("stop tracking container %s with id %s", containerName, containerID)
		track <- Track{ContainerID: containerID, Remove: true}
	}()
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	options := types.ContainerLogsOptions{ShowStdout: true, Follow: true, Tail: "1"}
	// Replace this ID with a container that really exists
	out, err := cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	scanner := bufio.NewReader(out)
	for {
		l, err := scanner.ReadString('\n')
		if err != nil {
			break
		}
		if !logRegex.MatchString(l) {
			log.Debugf("Log not match: %s", l)
			continue
		}

		match := logRegex.FindStringSubmatch(l)
		// labels := map[string]int{}
		lv := []string{containerName}
		for i, label := range logRegex.SubexpNames() {
			if i != 0 && label != "" {
				if strings.HasPrefix(label, "label_") {
					// labels[label] = i
					lv = append(lv, match[i])
				}
			}
		}

		// for _, v := range labels {
		// 	lv = append(lv, match[v])
		// }
		// log.Info("label list sort %+q", lv)
		for i, name := range logRegex.SubexpNames() {
			if i != 0 && name != "" {
				if strings.HasPrefix(name, "label_") {
					continue
				}

				switch m := metric[name].(type) {
				case *prometheus.CounterVec:
					v, err := strconv.ParseFloat(match[i], 64)
					if err != nil && !strings.HasPrefix(name, "optional_") {
						log.Warnf("Inlvaid %s float for %s (err: %s)", match[i], name, err.Error())
						continue
					}
					m.WithLabelValues(lv...).Add(v)
				case *prometheus.GaugeVec:
					v, err := strconv.ParseFloat(match[i], 64)
					if err != nil && !strings.HasPrefix(name, "optional_") {
						log.Warnf("Inlvaid %s float for %s (err: %s)", match[i], name, err.Error())
						continue
					}
					m.WithLabelValues(lv...).Set(v)
				case *prometheus.HistogramVec:
					v, err := strconv.ParseFloat(match[i], 64)
					if err != nil && !strings.HasPrefix(name, "optional_") {
						log.Warnf("Inlvaid %s float for %s (err: %s)", match[i], name, err.Error())
						continue
					}
					m.WithLabelValues(lv...).Observe(v)
					// m.Write(metric[name])
				default:
					log.Warnf("%T not supported", m)
				}
			}
		}
	}

}
