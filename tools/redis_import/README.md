## Notice

1. We don't support `set` data type.
2. Our `zset` use integer instead of double, so the zset float score in Redis 
   will be **converted to integer**.
3. Only Support Redis version greater than  `2.8.0`, because we use `scan` command to scan data.
   Also, you need `redis-py` greater than `2.9.0`

