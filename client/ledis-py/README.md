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

See [ledis wiki](https://github.com/siddontang/ledisdb/wiki/Commands) for more informations.


## Connection

### Connection Pools

### Connnections

