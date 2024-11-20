#### README
This API is designed to build up the IPQS plugin for Traefik and Caddy by [@z3ntl3](https://github.com/z3ntl3)

There are rules for it to be used effectively, with no additional costs, performance-vise.

##### Rules
- Always allocate no more than one ``ipqs.Client`` using ``ipqs.New``
- Reuse the same client using a reference, which should be allocated no more than once

### Performance | important
When using this package, the caller must either use the API with own caching implementation, or prefer to enable the default in-memory cache by setting ``ipqs.EnableCaching = true``

#### When using the built-in cache
- Set the TTL once by having a ``context.WithValue`` as a parent to your context, be sure to use ``ipqs.TTL_key`` to set  the key

#### Bench (with in memory cache enabled)
```
Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkClient$ github.com/SimpaiX-net/ipqs/tests

goos: darwin
goarch: arm64
pkg: github.com/SimpaiX-net/ipqs/tests
cpu: Apple M1
BenchmarkClient-8   	 3177540	       434.4 ns/op	     560 B/op	       7 allocs/op
PASS
ok  	github.com/SimpaiX-net/ipqs/tests	2.688s

```