#ledis-py

The Python interface to the ledisdb key-value store.


##Installation


ledis-py requires a running ledisdb server. See [ledisdb guide](https://github.com/siddontang/ledisdb#build-and-install) for installation instructions.

To install ledis-py, simply using `pip`(recommended):

```
$ sudo pip install ledis
```

or alternatively, using `easy_install`:

```
$ sudo easy_install ledis
```

or install from the source:

```
$ sudo python setup.py install 
```

##Getting Started

```
>>> import ledis
>>> l = ledis.Ledis(host='localhost', port=6380, db=0)
>>> l.set('foo', 'bar')
True
>>> l.get('foo')
'bar'
>>> 
```

## API Reference

For full API reference, please visit [rtfd](http://ledis-py.readthedocs.org/).


## Connection

### Connection Pools

Behind the scenes, ledis-py uses a connection pool to manage connections to a Ledis server. By default, each Ledis instance you create will in turn create its own connection pool. You can override this behavior and use an existing connection pool by passing an already created connection pool instance to the connection_pool argument of the Ledis class. You may choose to do this in order to implement client side sharding or have finer grain control of how connections are managed.

```
>>> pool = ledis.ConnectionPool(host='localhost', port=6380, db=0)
>>> l = ledis.Ledis(connection_pool=pool)
```

### Connections

ConnectionPools manage a set of Connection instances. ledis-py ships with two types of Connections. The default, Connection, is a normal TCP socket based connection. The UnixDomainSocketConnection allows for clients running on the same device as the server to connect via a unix domain socket. To use a UnixDomainSocketConnection connection, simply pass the unix_socket_path argument, which is a string to the unix domain socket file. Additionally, make sure the unixsocket parameter is defined in your `ledis.json` file. e.g.:

```
{
    "addr": "/tmp/ledis.sock",
    ...
}
```

```
>>> l = ledis.Ledis(unix_socket_path='/tmp/ledis.sock')
```

You can create your own Connection subclasses as well. This may be useful if you want to control the socket behavior within an async framework. To instantiate a client class using your own connection, you need to create a connection pool, passing your class to the connection_class argument. Other keyword parameters your pass to the pool will be passed to the class specified during initialization.

```
>>> pool = ledis.ConnectionPool(connection_class=YourConnectionClass,
                                your_arg='...', ...)
```

e.g.:

```
>>> from ledis import UnixDomainSocketConnection
>>> pool = ledis.ConnectionPool(connection_class=UnixDomainSocketConnection, path='/tmp/ledis.sock')
```

## Response Callbacks

The client class uses a set of callbacks to cast Ledis responses to the appropriate Python type. There are a number of these callbacks defined on the Ledis client class in a dictionary called RESPONSE_CALLBACKS.

Custom callbacks can be added on a per-instance basis using the `set_response_callback` method. This method accepts two arguments: a command name and the callback. Callbacks added in this manner are only valid on the instance the callback is added to. If you want to define or override a callback globally, you should make a subclass of the Ledis client and add your callback to its RESPONSE_CALLBACKS class dictionary.

Response callbacks take at least one parameter: the response from the Ledis server. Keyword arguments may also be accepted in order to further control how to interpret the response. These keyword arguments are specified during the command's call to execute_command. The ZRANGE implementation demonstrates the use of response callback keyword arguments with its "withscores" argument.