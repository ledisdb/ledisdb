## Notice

1. The tool doesn't support `set` data type.
2. The tool doesn't support `bitmap` data type.
2. Our `zset` use integer instead of double, so the zset float score in Redis 
   will be **converted to integer**.
3. Only Support Redis version greater than  `2.8.0`, because we use `scan` command to scan data.
   Also, you need `redis-py` greater than `2.9.0`.



## Usage


       $ python redis_import.py redis_host redis_port redis_db ledis_host ledis_port


We will use the same db index as redis. That's to say, data in redis[0] will be transfer to ledisdb[0]. But if redis db `index >= 16`, we will refuse to transfer, because ledisdb only support db `index < 16`.