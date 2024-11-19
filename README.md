#### README
Using this API requires to use caching in production to prevent hindering in service response times.
It is also better to set a rotating proxy as proxy for the client.

This package **comes with a default in-memory cache**. Use ``ipqs.EnableCaching`` to enable this default feature. Keep in mind that you should ever only just allocate one ``Client`` and re-use it through a reference. Do not allocate more clients than once.

