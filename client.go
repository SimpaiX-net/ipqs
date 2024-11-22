package ipqs

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	// InternetDB endpoint
	InternetDB = "https://internetdb.shodan.io/"

	// Control whether to enable caching or not
	EnableCaching = true

	DefaultTTL = time.Hour * 24 * 7
)

// IPQS client
type Client struct {
	proxy string
	httpClient    http.Client
	cache sync.Map // in memory cache
}

type CacheItem = struct {
	exp   int64
	score CacheIndex
}

type CacheIndex = uint8

type Result = uint8

// must be done to avoid
// collision with key of [context.WithValue]
type ttl string

const TTL_key ttl = "ttl"

const (
	// Good reputation
	GOOD Result = 1
	// Unknown reputation
	UNKNOWN Result = 2
	// Bad reputation
	BAD Result = 3
)

// Creates new IPQS client
func New() *Client {
	return new(Client)
}

func (c *Client) SetProxy(proxy string) *Client {
	if proxy != "" {
		c.proxy = proxy
	}
	
	return c
}


// Provisions the client
func (c *Client) Provision() error {
	if c.proxy != "" {
		uri, err := url.Parse(c.proxy)
		if err != nil {
			return err
		}
	
		c.httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(uri),
		}
	}
	
	return nil
}

// 	Gets the result for the ip query score [cached]
//
// 	query is the ip/hostname to query
//
//	userAgent will be used to set the request user agent
// 
//  The returned error can provide useful details
// 
//  nil means legit reputation, in constrast when it is not,
//  you should check for ErrBadIPRep and ErrUnknown, and any other error which usually
//  means a failure on reaching the InternetDB service
func (c *Client) GetIPQS(ctx context.Context, query, userAgent string) error {
	var cache CacheItem

	if EnableCaching {
		cache, hit := c.cache.Load(query)
		if hit {
			// cache hit
			cache := cache.(CacheItem)

			// check ttl expiration
			if time.Now().Unix() < cache.exp {
				if cache.score == BAD {
					return ErrBadIPRep
				} 
				
				if cache.score == UNKNOWN {
					return ErrUnknown
				}
				return nil
			}
		}
	}

	exp, ok := ctx.Value(TTL_key).(time.Duration)
	if !ok || exp == 0 {
		exp = DefaultTTL
	}

	cache.exp = time.Now().Add(exp).Unix()
	cache.score = UNKNOWN
	
	defer func() {
		// if we did defer c.cache.store(...) the store
		// would be evalauted immediately but we need
		// wait for adjustsments on it to be completed before
		// evaluating
		c.cache.Store(query, cache)
	}()
	

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, InternetDB + query, nil)
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Cache-Control", "must-revalidate")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(req); if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNotFound {
		cache.score = GOOD
		return nil
	}

	if res.StatusCode == http.StatusOK {
		cache.score = BAD
		return ErrBadIPRep
	}

	return ErrUnknown
}
