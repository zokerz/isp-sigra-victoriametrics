package routeros

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-routeros/routeros/v3"

	"mikrotik-victoriametrics-monitor/internal/config"
)

type Client interface {
	FetchInterfaces(ctx context.Context, router config.RouterConfig, filter config.FilterConfig) ([]InterfaceStats, int, error)
}

type RouterOSClient struct{}

func NewClient() *RouterOSClient {
	return &RouterOSClient{}
}

func (c *RouterOSClient) FetchInterfaces(ctx context.Context, r config.RouterConfig, filter config.FilterConfig) ([]InterfaceStats, int, error) {
	address := r.Address + ":" + strconv.Itoa(r.APIPort)

	type result struct {
		rows []map[string]string
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		var client *routeros.Client
		var err error
		if r.UseTLS {
			client, err = routeros.DialTLS(address, r.Username, r.Password, nil)
		} else {
			client, err = routeros.Dial(address, r.Username, r.Password)
		}
		if err != nil {
			ch <- result{err: err}
			return
		}
		defer client.Close()

		reply, err := client.Run("/interface/print", "=stats=")
		if err != nil {
			ch <- result{err: err}
			return
		}

		rows := make([]map[string]string, 0, len(reply.Re))
		for _, sentence := range reply.Re {
			row := make(map[string]string, len(sentence.Map))
			for k, v := range sentence.Map {
				row[k] = v
			}
			rows = append(rows, row)
		}
		ch <- result{rows: rows}
	}()

	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return nil, 0, res.err
		}
		stats := make([]InterfaceStats, 0, len(res.rows))
		for _, row := range res.rows {
			iface := ParseInterface(row)
			if IncludeInterface(iface.Name, filter) {
				stats = append(stats, iface)
			}
		}
		return stats, len(res.rows), nil
	case <-time.After(time.Hour):
		return nil, 0, fmt.Errorf("routeros client stalled")
	}
}
