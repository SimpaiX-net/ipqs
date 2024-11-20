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
	cache sync.Map // in memory cache
	fc    *fasthttp.Client
}

type CacheItem = struct {
	exp   int64
	score CacheIndex
}
type CacheIndex = uint8
type Result = uint8
type TTL = time.Duration

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
		fc: &fasthttp.Client{},
	}
}

// Sets proxy for use with IPQS client
// This must precede before GetIPQS() if you want it to use your proxy
func (c *Client) SetProxy(proxy string) *Client {
	c.proxy = proxy
	return c
}

// Finds the exact cause for query in the cache
//
// Compare the constants GOOD, BAD and UNKNOWN only when the returned
// bool is true, otherwise it means that there isn't a cache for the given query (key)
func (c *Client) FoundCause(query string) (Result, bool) {
	v, exists := c.cache.Load(query)
	score := v.(CacheItem).score

	return Result(score), exists
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
		c.fc.Dial = fasthttpproxy.
			FasthttpHTTPDialerDualStack(c.proxy)
	case "socks5":
		c.fc.Dial = fasthttpproxy.
			FasthttpSocksDialerDualStack(c.proxy)
	}

	return
}

// Gets the result for the ip query score
//
// query is the ip/hostname to query
//
//	userAgent will be used to set the request user agent
//
// This is a special sync function that will either return by cancellation signal,
// whether delegated by a cancel call to the given context, or timeout occuring, returning
// the associated error all together. If operation was successfull, the returned error will be nil.
//
// To find out the exact cause of any error, use the client.FoundCause method together with the constants:
// GOOD, BAD and UNKNOWN. Only compare those constants to the returned value when the returned boolean is true.
//
// It should not be checked when it is false, because that means that there isn't any cache for the given query aka key
func (c *Client) GetIPQS(ctx context.Context, query, userAgent string) error {
	ttl, ok := ctx.Value("ttl").(TTL)
	if !ok || ttl == 0 {
		ttl = TTL(time.Hour * 6)
	}

	done := make(chan error)
	go func() {
		store := CacheItem{}

		cache, hit := c.cache.Load(query)
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

		req.Header.Add("User-Agent", userAgent)
		req.Header.Add("Cache-Control", "must-revalidate")
		req.Header.Add("Content-Type", "application/json")

		req.SetConnectionClose()

		if err := c.fc.Do(req, res); err != nil {
			done <- err
			return
		}

		defer c.cache.Store(query, store)

		store.exp = time.Now().Add(time.Hour * 6).Unix()

		if res.StatusCode() != http.StatusOK &&
			res.StatusCode() != http.StatusNotFound {

			store.score = UNKNOWN
			done <- ErrUnknown
			return
		}

		if res.StatusCode() != http.StatusNotFound {
			store.score = BAD
			done <- ErrBadIPRep

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
