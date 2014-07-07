--- refer from openresty redis lib

local sub = string.sub
local byte = string.byte
local tcp = ngx.socket.tcp
local concat = table.concat
local null = ngx.null
local pairs = pairs
local unpack = unpack
local setmetatable = setmetatable
local tonumber = tonumber
local error = error


local ok, new_tab = pcall(require, "table.new")
if not ok then
    new_tab = function (narr, nrec) return {} end
end


local _M = new_tab(0, 155)
_M._VERSION = '0.01'


local commands = {
    --[[kv]]
    "decr",
    "decrby",
    "del",
    "exists",
    "get",
    "getset",
    "incr",
    "incrby",
    "mget",
    "mset",
    "set",
    "setnx",
    "ttl",
    "expire",
    "expireat",
    "persist",

    --[[hash]]
    "hdel", 
    "hexists", 
    "hget", 
    "hgetall", 
    "hincrby", 
    "hkeys", 
    "hlen", 
    "hmget", 
    --[["hmset",]] 
    "hset", 
    "hvals",
    --[[ledisdb special commands]] 
    "hclear", 
    "hmclear",
    "hexpire",
    "hexpireat",
    "httl",
    "hpersist",

    --[[list]]
    "lindex", 
    "llen", 
    "lpop", 
    "lpush", 
    "lrange", 
    "rpop", 
    "rpush", 
    --[[ledisdb special commands]]
    "lclear", 
    "lmclear",
    "lexpire",
    "lexpireat",
    "lttl",
    "lpersist",

    --[[zset]]
    "zadd", 
    "zcard", 
    "zcount", 
    "zincrby", 
    "zrange", 
    "zrangebyscore", 
    "zrank", 
    "zrem", 
    "zremrangebyrank", 
    "zremrangebyscore", 
    "zrevrange", 
    "zrevrank", 
    "zrevrangebyscore", 
    "zscore", 
    --[[ledisdb special commands]]
    "zclear", 
    "zmclear",
    "zexpire",
    "zexpireat",
    "zttl",
    "zpersist",

    --[[server]]
    "ping",
    "echo",
    "select"
}


local mt = { __index = _M }


function _M.new(self)
    local sock, err = tcp()
    if not sock then
        return nil, err
    end
    return setmetatable({ sock = sock }, mt)
end


function _M.set_timeout(self, timeout)
    local sock = self.sock
    if not sock then
        return nil, "not initialized"
    end

    return sock:settimeout(timeout)
end


function _M.connect(self, ...)
    local sock = self.sock
    if not sock then
        return nil, "not initialized"
    end

    return sock:connect(...)
end


function _M.set_keepalive(self, ...)
    local sock = self.sock
    if not sock then
        return nil, "not initialized"
    end

    return sock:setkeepalive(...)
end


function _M.get_reused_times(self)
    local sock = self.sock
    if not sock then
        return nil, "not initialized"
    end

    return sock:getreusedtimes()
end


local function close(self)
    local sock = self.sock
    if not sock then
        return nil, "not initialized"
    end

    return sock:close()
end
_M.close = close


local function _read_reply(self, sock)
    local line, err = sock:receive()
    if not line then
        if err == "timeout"  then
            sock:close()
        end
        return nil, err
    end

    local prefix = byte(line)

    if prefix == 36 then    -- char '$'
        -- print("bulk reply")

        local size = tonumber(sub(line, 2))
        if size < 0 then
            return null
        end

        local data, err = sock:receive(size)
        if not data then
            if err == "timeout" then
                sock:close()
            end
            return nil, err
        end

        local dummy, err = sock:receive(2) -- ignore CRLF
        if not dummy then
            return nil, err
        end

        return data

    elseif prefix == 43 then    -- char '+'
        -- print("status reply")

        return sub(line, 2)

    elseif prefix == 42 then -- char '*'
        local n = tonumber(sub(line, 2))

        -- print("multi-bulk reply: ", n)
        if n < 0 then
            return null
        end

        local vals = new_tab(n, 0);
        local nvals = 0
        for i = 1, n do
            local res, err = _read_reply(self, sock)
            if res then
                nvals = nvals + 1
                vals[nvals] = res

            elseif res == nil then
                return nil, err

            else
                -- be a valid redis error value
                nvals = nvals + 1
                vals[nvals] = {false, err}
            end
        end

        return vals

    elseif prefix == 58 then    -- char ':'
        -- print("integer reply")
        return tonumber(sub(line, 2))

    elseif prefix == 45 then    -- char '-'
        -- print("error reply: ", n)

        return false, sub(line, 2)

    else
        return nil, "unkown prefix: \"" .. prefix .. "\""
    end
end


local function _gen_req(args)
    local nargs = #args

    local req = new_tab(nargs + 1, 0)
    req[1] = "*" .. nargs .. "\r\n"
    local nbits = 1

    for i = 1, nargs do
        local arg = args[i]
        nbits = nbits + 1

        if not arg then
            req[nbits] = "$-1\r\n"

        else
            if type(arg) ~= "string" then
                arg = tostring(arg)
            end
            req[nbits] = "$" .. #arg .. "\r\n" .. arg .. "\r\n"
        end
    end

    -- it is faster to do string concatenation on the Lua land
    return concat(req)
end


local function _do_cmd(self, ...)
    local args = {...}

    local sock = self.sock
    if not sock then
        return nil, "not initialized"
    end

    local req = _gen_req(args)

    local reqs = self._reqs
    if reqs then
        reqs[#reqs + 1] = req
        return
    end

    -- print("request: ", table.concat(req))

    local bytes, err = sock:send(req)
    if not bytes then
        return nil, err
    end

    return _read_reply(self, sock)
end




function _M.read_reply(self)
    local sock = self.sock
    if not sock then
        return nil, "not initialized"
    end

    local res, err = _read_reply(self, sock)

    return res, err
end


for i = 1, #commands do
    local cmd = commands[i]

    _M[cmd] =
        function (self, ...)
            return _do_cmd(self, cmd, ...)
        end
end


function _M.hmset(self, hashname, ...)
    local args = {...}
    if #args == 1 then
        local t = args[1]

        local n = 0
        for k, v in pairs(t) do
            n = n + 2
        end

        local array = new_tab(n, 0)

        local i = 0
        for k, v in pairs(t) do
            array[i + 1] = k
            array[i + 2] = v
            i = i + 2
        end
        -- print("key", hashname)
        return _do_cmd(self, "hmset", hashname, unpack(array))
    end

    -- backwards compatibility
    return _do_cmd(self, "hmset", hashname, ...)
end


function _M.array_to_hash(self, t)
    local n = #t
    -- print("n = ", n)
    local h = new_tab(0, n / 2)
    for i = 1, n, 2 do
        h[t[i]] = t[i + 1]
    end
    return h
end


function _M.add_commands(...)
    local cmds = {...}
    for i = 1, #cmds do
        local cmd = cmds[i]
        _M[cmd] =
            function (self, ...)
                return _do_cmd(self, cmd, ...)
            end
    end
end


return _M