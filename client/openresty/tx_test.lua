local ledis = require "ledis"
local lds = ledis:new()

lds:set_timeout(1000)




-- connect
local ok, err =  lds:connect("127.0.0.1", "6380")
if not ok then
	ngx.say("failed to connect:", err)
	return
end

lds:del("tx")

-- transaction

ok, err = lds:set("tx", "a")
if not ok then
	ngx.say("failed to execute set in tx: ", err)
	return
end

ngx.say("SET should be OK <=>", ok)

res, err = lds:get("tx")
if not res then
	ngx.say("failed to execute get in tx: ", err)
	return
end

ngx.say("GET should be a <=>", res)



ok, err = lds:begin()
if not ok then
	ngx.say("failed to run begin: ", err)
	return
end

ngx.say("BEGIN should be OK <=>", ok)

ok, err = lds:set("tx", "b")
if not ok then
	ngx.say("failed to execute set in tx: ", err)
	return
end

ngx.say("SET should be OK <=>", ok)


res, err = lds:get("tx")
if not res then
	ngx.say("failed to execute get in tx: ", err)
	return
end

ngx.say("GET should be b <=>", res)

ok, err = lds:rollback()
if not ok then
	ngx.say("failed to rollback", err)
	return
end
ngx.say("ROLLBACK should be OK <=>", ok)

res, err = lds:get("tx")
if not res then
	ngx.say("failed to execute get in tx: ", err)
	return
end

ngx.say("GET should be a <=>", res)


lds:begin()
lds:set("tx", "c")
lds:commit()
res, err = lds:get("tx")
if not res then
	ngx.say("failed to execute get in tx: ", err)
	return
end

ngx.say("GET should be c <=>", res)


local ok, err = lds:close()
if not ok then
    ngx.say("failed to close: ", err)
    return
end
ngx.say("close success")
