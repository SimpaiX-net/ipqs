/*
todo..........
api might change its not final
*/
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

type CacheItem = struct {
	exp   int64
	score CacheIndex
}
type CacheIndex = uint8
type Result = uint8

const (
	// Good reputation
	GOOD Result = 0
	// Unknown reputation
	UNKNOWN Result = 1
	// Bad reputation
	BAD Result = 2
)

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

// Finds the exact cause for query in the cache
// Result returns either GOOD, BAD or UNKNOWN
func (c *Client) Cause(query string) (Result, bool) {
	v, exists := c.Map.Load(query)
	score := v.(CacheItem).score

	return Result(score), exists
}

// Provisions the client
func (c *Client) Provision() (err error) {
	if c.ttl == 0 {
		c.ttl = time.Hour * 6
	}

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

// Gets the result for the ip query score
//
// query is the ip/hostname to query
//
//	user_agent will be used to set the request user agent
//
// If this function returns and error then, the query had either a bad reputation
// or it was unknown.
//
// To find out the cause you can use client.Cause(query) which returns the result.
//
// Use constants like BAD, GOOD or UNKNOWN to check against the result client.Cause returns
func (c *Client) GetIPQS(ctx context.Context, query, user_agent string, done chan error) error {
	// apply timeout/ctx cancellation signal to the whole functionality
	// the operation can complete with success before any of that occurs
	//
	// cancel must be called to free up resources after this method returns
	go func() {
		store := CacheItem{}

		cache, hit := c.Map.Load(query)
		if hit {
			// cache hit
			cache := cache.(CacheItem)

			// check ttl expiration
			if time.Now().Unix() < cache.exp {
				if cache.score == BAD {
					done <- ErrBadIPRep
					return
				} else if cache.score == UNKNOWN {
					done <- ErrUnknown
					return
				}

				done <- nil
				return
			}
		}

		req := fasthttp.AcquireRequest()
		res := fasthttp.AcquireResponse()

		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(res)

		req.SetRequestURI(InternetDB + query)

		req.Header.Add("User-Agent", user_agent)
		req.Header.Add("Cache-Control", "must-revalidate")
		req.Header.Add("Content-Type", "application/json")

		req.SetConnectionClose()

		if err := c.Do(req, res); err != nil {
			done <- err
			return
		}

		defer c.Map.Store(query, store)

		store.exp = time.Now().Add(c.ttl).Unix()

		if res.StatusCode() != http.StatusOK &&
			res.StatusCode() != http.StatusNotFound {

			store.score = UNKNOWN
			done <- ErrUnknown
			return
		}

		if res.StatusCode() != http.StatusNotFound {
			done <- ErrBadIPRep
			store.score = BAD

			return
		}

		store.score = GOOD
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
