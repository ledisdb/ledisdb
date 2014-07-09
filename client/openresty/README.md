Name
=====

lua-resty-ledis - Lua ledisdb client driver for the ngx_lua based  on the cosocket API



Status
======

The library is not production ready. Do it at your own risk.


Description
===========

This Lua library is a [ledisdb](https://github.com/siddontang/ledisdb) client driver for the ngx_lua nginx module:

http://wiki.nginx.org/HttpLuaModule

This Lua library takes advantage of ngx_lua's cosocket API, which ensures 100% nonblocking behavior.

Note that at least [ngx_lua 0.5.14](https://github.com/openresty/lua-nginx-module/tags) or [ngx_openresty 1.2.1.14](http://openresty.org/#Download) is required.


Synopsis
========

```lua
  	# you have to add the following line
    lua_package_path "/path/to/ledis/client/openresty/?.lua;;";

    server {
        location /test {
            content_by_lua '
                local ledis = require "ledis"
                local lds = ledis:new()

                lds:set_timeout(1000) -- 1 sec

                -- or connect to a unix domain socket file listened
                -- by a ledis server:
                --     local ok, err = lds:connect("unix:/path/to/ledis.sock")

                local ok, err = lds:connect("127.0.0.1", 6380)
                if not ok then
                    ngx.say("failed to connect: ", err)
                    return
                end

                ok, err = lds:set("dog", "an animal")
                if not ok then
                    ngx.say("failed to set dog: ", err)
                    return
                end

                ngx.say("set result: ", ok)

                local res, err = lds:get("dog")
                if not res then
                    ngx.say("failed to get dog: ", err)
                    return
                end

                if res == ngx.null then
                    ngx.say("dog not found.")
                    return
                end

                ngx.say("dog: ", res)

                -- put it into the connection pool of size 100,
                -- with 10 seconds max idle time
                local ok, err = lds:set_keepalive(10000, 100)
                if not ok then
                    ngx.say("failed to set keepalive: ", err)
                    return
                end

                -- or just close the connection right away:
                -- local ok, err = lds:close()
                -- if not ok then
                --     ngx.say("failed to close: ", err)
                --     return
                -- end
            ';
        }
    }

```



Methods
=========

All of the ledisdb commands have their own methods with the same name except all in **lower case**.

You can find the complete list of ledisdb commands here"

https://github.com/siddontang/ledisdb/wiki/Commands

You need to check out this ledisdb command reference to see what ledisdb command accepts what arguments.

The ledisdb command arguments can be directly fed into the corresponding method call. For example, the `GET` ledisdb command accepts a single key argument, then you can just call the `get` method like this:


    local res, err = lds:get("key")

Similarly, the "LRANGE" ledisdb command accepts threee arguments, then you should call the "lrange" method like this:

    local res, err = lds:lrange("nokey", 0, 1)

For example, "SET", "GET", "LRANGE", and "LPOP" commands correspond to the methods "set", "get", "lrange", and "lpop".

Here are some more examples:

    -- HMGET myhash field1 field2 nofield
    local res, err = lds:hmget("myhash", "field1", "field2", "nofield")
    -- HMSET myhash field1 "Hello" field2 "World"
    local res, err = lds:hmset("myhash", "field1", "Hello", "field2", "World")

All these command methods returns a single result in success and nil otherwise. In case of errors or failures, it will also return a second value which is a string describing the error.

All these command methods returns a single result in success and nil otherwise. In case of errors or failures, it will also return a second value which is a string describing the error.

- A Redis "status reply" results in a string typed return value with the "+" prefix stripped.

- A Redis "integer reply" results in a Lua number typed return value.

- A Redis "error reply" results in a false value and a string describing the error.

- A non-nil Redis "bulk reply" results in a Lua string as the return value. A nil bulk reply results in a ngx.null return value.

- A non-nil Redis "multi-bulk reply" results in a Lua table holding all the composing values (if any). If any of the composing value is a valid redis error value, then it will be a two element table {false, err}.

- A nil multi-bulk reply returns in a ngx.null value.

See http://redis.io/topics/protocol for details regarding various Redis reply types.

In addition to all those ledisdb command methods, the following methods are also provided:


new
====

	synxtax: lds, err = ledis:new()

Creates a ledis object. In case of failures, returns nil and a string describing the error.


connect
========

    syntax: ok, err = lds:connect(host, port, options_table?)

	syntax: ok, err = lds:connect("unix:/path/to/unix.sock", options_table?)


Attempts to connect to the remote host and port that the ledis server is listening to or a local unix domain socket file listened by the ledis server.

Before actually resolving the host name and connecting to the remote backend, this method will always look up the connection pool for matched idle connections created by previous calls of this method.

An optional Lua table can be specified as the last argument to this method to specify various connect options:

- pool

Specifies a custom name for the connection pool being used. If omitted, then the connection pool name will be generated from the string template ` <host>:<port>` or `<unix-socket-path>`.



set_timeout
============

	syntax: lds:set_timeout(time)

Sets the timeout (in `ms`) protection for subsequent operations, including the connect method.


set_keepalive
==============

	syntax: ok, err = lds:set_keepalive(max_idle_timeout, pool_size)

Puts the current ledis connection immediately into the ngx_lua cosocket connection pool.

You can specify the max idle timeout (in ms) when the connection is in the pool and the maximal size of the pool every nginx worker process.

In case of success, returns 1. In case of errors, returns nil with a string describing the error.

Only call this method in the place you would have called the close method instead. Calling this method will immediately turn the current ledis object into the `closed` state. Any subsequent operations other than connect() on the current objet will return the closed error.


get_reused_times
=================

	syntax: times, err = lds:get_reused_times()

This method returns the (successfully) reused times for the current connection. In case of error, it returns nil and a string describing the error.

If the current connection does not come from the built-in connection pool, then this method always returns 0, that is, the connection has never been reused (yet). If the connection comes from the connection pool, then the return value is always non-zero. So this method can also be used to determine if the current connection comes from the pool.


close
=======

	syntax: ok, err = lds:close()

Closes the current ledis connection and returns the status.

In case of success, returns 1. In case of errors, returns nil with a string describing the error.



hmset
======


	syntax: lds:hmset(myhash, field1, value1, field2, value2, ...)

	syntax: lds:hmset(myhash, { field1 = value1, field2 = value2, ... })

Special wrapper for the ledis `hmset` command.

When there are only three arguments (including the "lds" object itself), then the last argument must be a Lua table holding all the field/value pairs.


add_commands
============

	syntax: hash = ledis.add_commands(cmd_name1, cmd_name2, ...)

Adds new ledis commands to the resty.ledis class. Here is an example:

```lua
    local ledis = require "ledis"

    ledis.add_commands("foo", "bar")

    local lds = ledis:new()

    lds:set_timeout(1000) -- 1 sec	

    local ok, err = lds:connect("127.0.0.1", 6380)
    if not ok then
        ngx.say("failed to connect: ", err)
        return
    end

    local res, err = lds:foo("a")
    if not res then
        ngx.say("failed to foo: ", err)
    end

    res, err = lds:bar()
    if not res then
        ngx.say("failed to bar: ", err)
    end

```


Debugging
=========

It is usually convenient to use the lua-cjson library to encode the return values of the ledis command methods to JSON. For example,

    local cjson = require "cjson"
    ...
    local res, err = lds:mget("h1234", "h5678")
    if res then
        print("res: ", cjson.encode(res))
    end


Automatic Error Logging
========================


By default the underlying [ngx_lua](http://wiki.nginx.org/HttpLuaModule) module does error logging when socket errors happen. If you are already doing proper error handling in your own Lua code, then you are recommended to disable this automatic error logging by turning off ngx_lua's [lua_socket_log_errors](http://wiki.nginx.org/HttpLuaModule#lua_socket_log_errors) directive, that is,

    lua_socket_log_errors off;


Check List for Issues
=======================

Please refer to [lua-resty-redis](https://github.com/openresty/lua-resty-redis#check-list-for-issues).


Limitations
===========

Please refer to [lua-resty-redis](https://github.com/openresty/lua-resty-redis#limitations).


Installation
============
 
 You need to configure the `lua_package_path` directive to add the path of your lua-resty-ledis source tree to ngx_lua's `LUA_PATH` search path, as in

    # nginx.conf
    http {
        lua_package_path "/path/to/ledis/client/openresty/?.lua;;";
        ...
    }
Ensure that the system account running your Nginx ''worker'' proceses have enough permission to read the `.lua` file.



Bugs and Patches
================

Please report bugs or submit patches by [shooting an issue](https://github.com/siddontang/ledisdb/issues/new). Thank you.


Author
======

The original author is Yichun "agentzh" Zhang (章亦春) agentzh@gmail.com, CloudFlare Inc.


Thanks
======

Thanks Yichun "agentzh" Zhang (章亦春) for making such great works.