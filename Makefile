.PHONY: all clean

all: haproxy_abuser_exporter

clean:
	rm -f haproxy_abuser_exporter

haproxy_abuser_exporter: main.go Scraper.go
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o haproxy_abuser_exporter
