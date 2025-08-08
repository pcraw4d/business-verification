package datasource

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

// EnrichmentRequest carries fields used to look up external business data
type EnrichmentRequest struct {
	BusinessName       string
	RegistrationNumber string
	CountryCode        string
}

// EnrichmentResult is a normalized bundle of optional data used to improve classification
type EnrichmentResult struct {
	CleanBusinessName string
	Industry          string
	Description       string
	Keywords          []string
}

// DataSource defines a pluggable enrichment source (DB, 3rd-party API, etc.)
type DataSource interface {
	Name() string
	Enrich(ctx context.Context, req EnrichmentRequest) (EnrichmentResult, error)
	HealthCheck(ctx context.Context) error
}

// Aggregator queries multiple sources with a timeout and merges results
type Aggregator struct {
	sources []DataSource
	timeout time.Duration
	client  *http.Client
}

// NewAggregator constructs a new aggregator
func NewAggregator(sources []DataSource, timeout time.Duration) *Aggregator {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &Aggregator{sources: sources, timeout: timeout}
}

// SetHTTPClient installs a shared pooled HTTP client for use by data sources that perform HTTP calls.
func (a *Aggregator) SetHTTPClient(client *http.Client) {
	a.client = client
}

// NewPooledHTTPClient creates a pooled HTTP client based on provided settings.
func NewPooledHTTPClient(maxIdleConns, maxIdlePerHost int, idleTimeout, tlsHandshakeTimeout, expectContinueTimeout, requestTimeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdlePerHost,
		IdleConnTimeout:       idleTimeout,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ExpectContinueTimeout: expectContinueTimeout,
	}
	return &http.Client{Transport: transport, Timeout: requestTimeout}
}

// Enrich queries all sources concurrently, returns a merged result
func (a *Aggregator) Enrich(ctx context.Context, req EnrichmentRequest) (EnrichmentResult, error) {
	if len(a.sources) == 0 {
		return EnrichmentResult{}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	type out struct {
		res EnrichmentResult
		err error
	}

	ch := make(chan out, len(a.sources))
	var wg sync.WaitGroup
	wg.Add(len(a.sources))
	for _, src := range a.sources {
		src := src
		go func() {
			defer wg.Done()
			res, err := src.Enrich(ctx, req)
			select {
			case ch <- out{res: res, err: err}:
			case <-ctx.Done():
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	merged := EnrichmentResult{}
	for o := range ch {
		if o.err != nil {
			continue
		}
		if merged.CleanBusinessName == "" && o.res.CleanBusinessName != "" {
			merged.CleanBusinessName = o.res.CleanBusinessName
		}
		if merged.Industry == "" && o.res.Industry != "" {
			merged.Industry = o.res.Industry
		}
		if merged.Description == "" && o.res.Description != "" {
			merged.Description = o.res.Description
		}
		if len(o.res.Keywords) > 0 {
			// append unique keywords
			exist := make(map[string]struct{}, len(merged.Keywords))
			for _, k := range merged.Keywords {
				exist[k] = struct{}{}
			}
			for _, k := range o.res.Keywords {
				if _, ok := exist[k]; !ok {
					merged.Keywords = append(merged.Keywords, k)
					exist[k] = struct{}{}
				}
			}
		}
	}
	return merged, nil
}
