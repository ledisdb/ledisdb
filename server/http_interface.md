##HTTP Interface
LedisDB provides http interfaces for most commands.
####Request
The proper url format is

    http://host:port[/db]/cmd/arg1/arg2/.../argN[?type=type]

'db' and 'type' are optional. 'db' stands for ledis db index, ranges from 0 to 15, its default value is 0. 'type' is a custom content type, can be json, bson or msgpack,  json is default.


####Response

The response format is
    
    { cmd: return_value }

or

    { cmd: [success, message] }

'return_value' stands for the output of 'cmd', it can be a number, a string, a list, or a hash. If the return value  is just a descriptive message, the second format will be taken, and 'success', a boolean value,  indicates whether it is successful. 

####Example
#####Curl

    curl http://127.0.0.1:11181/SET/hello/world
    → {"SET":[true,"OK"]}

    curl http://127.0.0.1:11181/0/GET/hello?type=json
    → {"GET":"world"}

#####Python
Requires [msgpack-python](https://pypi.python.org/pypi/msgpack-python) and [requests](https://pypi.python.org/pypi/requests/)    
    
    >>> import requests
    >>> import msgpack
    
    >>> requests.get("http://127.0.0.1:11181/0/SET/hello/world")
    >>> r = requests.get("http://127.0.0.1:11181/0/GET/hello?type=msgpack")
    >>> msgpack.unpackb(r.content) 
    >>> {"GET":"world"}
    
