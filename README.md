# Logprom

Config base prometheus exporter for logs, written in Go.
## Roadmap
### Prometheus Metrics
- Supported Metrics:
	- Counter
	- Guage
 - Will support:
    - Histogram
    - Summary
### Log input
Currently logprom only get data from from docker logs but it will support file, stdout and maybe gelft in future.
  

## Quick Start

### Run with docker

- Create your config file `logprom.yml` (for more detail refer to [config file](#config-file) section)

```
---
logprom:
- container:
  name: "test.*"
  log_format: "\\[(?P<label_platform>.+)\\] \\[metrics\\] \\{request_duration\\: (?P<request_duration>\\d+(\\.\\d+)?),( dummy\\: (?P<optional_dummy>\\d+(\\.\\d+)?),)? total_declared: (?P<total_declared>\\d+)"
  prometheus_mapping:
    request_duration: gauge
    total_declared: counter
    optional_dummy: counter
```
- Build docker image
```
docker build -t logprom .
```
- Then run dokcer container for logprom
```
docker run -d --name=logprom \
  -p 3030:3030 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v logprom.yml:/etc/logprom.yml \
  logprom
```
- Check your metrics at ‍`‍http://127.0.0.1:3030/metrics`‍‍

# Config File
Will be completed soon!
