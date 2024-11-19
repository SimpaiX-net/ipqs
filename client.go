package ipqs

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

var (
	// InternetDB endpoint
	InternetDB = "https://internetdb.shodan.io/"

	// If you want to enable internal caching
	EnableCaching = false
)

// IPQS client
type Client struct {
	proxy string
	ttl   time.Duration
	sync.Map
	*fasthttp.Client
}

type CacheItem = string

// Creates new IPQS client
func New() *Client {
	return &Client{
		Client: &fasthttp.Client{},
	}
}

// Sets the time to live for the cache
func (c *Client) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// Sets proxy for use with IPQS client
// This must precede before GetIPQS() if you want it to use your proxy
func (c *Client) SetProxy(proxy string) *Client {
	c.proxy = proxy
	return c
}

// Provisions the client
func (c *Client) Provision() (err error) {
	if c.proxy == "" {
		return
	}

	uri, err := url.Parse(c.proxy)
	if err != nil {
		return
	}

	var ok bool
	for _, scheme := range supportedProtocols {
		if scheme == uri.Scheme {
			ok = true
		}
	}

	if !ok {
		return ErrInvalidProtocol
	}

	switch uri.Scheme {
	case "http", "https":
		c.Dial = fasthttpproxy.
			FasthttpHTTPDialerDualStack(c.proxy)
	case "socks5":
		c.Dial = fasthttpproxy.
			FasthttpSocksDialerDualStack(c.proxy)
	}

	return
}

// Gets the IP Quality scan results
//
// This will send a request towards InternetDB with lookup as parameter
// to identify the trust score, determined by InternetDB
//
//	Given user_agent is set for the request
//
// This will either timeout using c.ctx, set by c.Provision(ctx) or finalize it's task with absolute success
func (c *Client) GetIPQS(ctx context.Context, lookup, user_agent string) error {
	cache, hit := c.Map.Load(lookup)
	if hit {
		if time.Now().Unix() <= cache.(int64) {
			return nil
		}

		// TTL is expired
		c.Map.Delete(lookup)
	}

	done := make(chan error)

	go func() {
		req := fasthttp.AcquireRequest()
		res := fasthttp.AcquireResponse()

		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(res)

		req.SetRequestURI(InternetDB + lookup)

		req.Header.Add("User-Agent", user_agent)
		req.Header.Add("Cache-Control", "must-revalidate")
		req.Header.Add("Content-Type", "application/json")

		req.SetConnectionClose()

		if err := c.Do(req, res); err != nil {
			done <- err
			return
		}

		defer c.Map.Store(lookup, time.Now().Add(c.ttl).Unix())

		if res.StatusCode() != http.StatusNotFound {
			done <- ErrBadIPRep
			return
		}

		done <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
