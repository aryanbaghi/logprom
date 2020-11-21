package logprom

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
)

func TestLog(t *testing.T) {

	cid := startContainer()
	rpcDurations := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "total_declared"}, []string{})
	r := regexp.MustCompile(`.*\[metrics\] .* \"total_declared\"\: (?P<total_declared>\d+).*`)
	a := map[string]interface{}{
		"total_declared": rpcDurations,
	}
	go trackContainer(cid, "test", r, a)
	time.Sleep(5 * time.Second)
	stopContainer(cid)
}

func stopContainer(cid string) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	err = cli.ContainerStop(ctx, cid, nil)
	if err != nil {
		panic(err)
	}
	err = cli.ContainerRemove(ctx, cid, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}
}

func startContainer() string {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	imageName := "busybox"

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   strslice.StrSlice{`sh`, `-c`, `while true; do echo 2020-03-03T08:38:22 [telegram] [metrics] {\"request_duration\": 0.956103, \"total_declared\": 61, \"requests\": 1, \"processed_contents\": 61, \"contents_with_problem\": 0, \"new_contents\": 61, \"old_contents\": 0}; sleep 1; done`},
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return resp.ID
}
