from __future__ import with_statement
import datetime
import time as mod_time
from ledis._compat import (b, izip, imap, iteritems, iterkeys, itervalues,
                           basestring, long, nativestr, urlparse, bytes)
from ledis.connection import ConnectionPool, UnixDomainSocketConnection
from ledis.exceptions import (
    ConnectionError,
    DataError,
    LedisError,
    ResponseError,
    ExecAbortError,
)

SYM_EMPTY = b('')


def list_or_args(keys, args):
    # returns a single list combining keys and args
    try:
        iter(keys)
        # a string or bytes instance can be iterated, but indicates
        # keys wasn't passed as a list
        if isinstance(keys, (basestring, bytes)):
            keys = [keys]
    except TypeError:
        keys = [keys]
    if args:
        keys.extend(args)
    return keys


def string_keys_to_dict(key_string, callback):
    return dict.fromkeys(key_string.split(), callback)


def dict_merge(*dicts):
    merged = {}
    [merged.update(d) for d in dicts]
    return merged


def pairs_to_dict(response):
    "Create a dict given a list of key/value pairs"
    it = iter(response)
    return dict(izip(it, it))


def zset_score_pairs(response, **options):
    """
    If ``withscores`` is specified in the options, return the response as
    a list of (value, score) pairs
    """
    if not response or not options['withscores']:
        return response
    score_cast_func = options.get('score_cast_func', int)
    it = iter(response)
    return list(izip(it, imap(score_cast_func, it)))


def int_or_none(response):
    if response is None:
        return None
    return int(response)


class Ledis(object):
    """
    Implementation of the Redis protocol.

    This abstract class provides a Python interface to all LedisDB commands
    and an implementation of the Redis protocol.

    Connection and Pipeline derive from this, implementing how
    the commands are sent and received to the Ledis server
    """
    RESPONSE_CALLBACKS = dict_merge(
        string_keys_to_dict(
            'EXISTS EXPIRE EXPIREAT HEXISTS HMSET SETNX '
            'PERSIST HPERSIST LPERSIST ZPERSIST',
            bool
        ),
        string_keys_to_dict(
            'DECRBY DEL HDEL HLEN INCRBY LLEN '
            'ZADD ZCARD ZREM ZREMRANGEBYRANK ZREMRANGEBYSCORE'
            'LMCLEAR HMCLEAR ZMCLEAR',
            int
        ),
        string_keys_to_dict(
            'LPUSH RPUSH',
            lambda r: isinstance(r, long) and r or nativestr(r) == 'OK'
        ),
        string_keys_to_dict(
            'MSET SELECT ',
            lambda r: nativestr(r) == 'OK'
        ),
        string_keys_to_dict(
            'ZRANGE ZRANGEBYSCORE ZREVRANGE ZREVRANGEBYSCORE',
            zset_score_pairs
        ),
        string_keys_to_dict('ZRANK ZREVRANK ZSCORE ZINCRBY', int_or_none),
        {
            'HGETALL': lambda r: r and pairs_to_dict(r) or {},
            'PING': lambda r: nativestr(r) == 'PONG',
            'SET': lambda r: r and nativestr(r) == 'OK', }
    )

    @classmethod
    def from_url(cls, url, db=None, **kwargs):
        """
        Return a Ledis client object configured from the given URL.

        For example::

            redis://[:password]@localhost:6380/0
            unix://[:password]@/path/to/socket.sock?db=0

        There are several ways to specify a database number. The parse function
        will return the first specified option:
            1. A ``db`` querystring option, e.g. redis://localhost?db=0
            2. If using the redis:// scheme, the path argument of the url, e.g.
               redis://localhost/0
            3. The ``db`` argument to this function.

        If none of these options are specified, db=0 is used.

        Any additional querystring arguments and keyword arguments will be
        passed along to the ConnectionPool class's initializer. In the case
        of conflicting arguments, querystring arguments always win.
        """
        connection_pool = ConnectionPool.from_url(url, db=db, **kwargs)
        return cls(connection_pool=connection_pool)

    def __init__(self, host='localhost', port=6380,
                 db=0, socket_timeout=None,
                 connection_pool=None, charset='utf-8',
                 errors='strict', decode_responses=False,
                 unix_socket_path=None):
        if not connection_pool:
            kwargs = {
                'db': db,
                'socket_timeout': socket_timeout,
                'encoding': charset,
                'encoding_errors': errors,
                'decode_responses': decode_responses,
            }
            # based on input, setup appropriate connection args
            if unix_socket_path:
                kwargs.update({
                    'path': unix_socket_path,
                    'connection_class': UnixDomainSocketConnection
                })
            else:
                kwargs.update({
                    'host': host,
                    'port': port
                })
            connection_pool = ConnectionPool(**kwargs)
        self.connection_pool = connection_pool
        self.response_callbacks = self.__class__.RESPONSE_CALLBACKS.copy()

    def set_response_callback(self, command, callback):
        "Set a custom Response Callback"
        self.response_callbacks[command] = callback

    #### COMMAND EXECUTION AND PROTOCOL PARSING ####

    def execute_command(self, *args, **options):
        "Execute a command and return a parsed response"
        pool = self.connection_pool
        command_name = args[0]
        connection = pool.get_connection(command_name, **options)
        try:
            connection.send_command(*args)
            return self.parse_response(connection, command_name, **options)
        except ConnectionError:
            connection.disconnect()
            connection.send_command(*args)
            return self.parse_response(connection, command_name, **options)
        finally:
            pool.release(connection)

    def parse_response(self, connection, command_name, **options):
        "Parses a response from the Ledis server"
        response = connection.read_response()
        if command_name in self.response_callbacks:
            return self.response_callbacks[command_name](response, **options)
        return response


    #### SERVER INFORMATION ####
    def echo(self, value):
        "Echo the string back from the server"
        return self.execute_command('ECHO', value)

    def ping(self):
        "Ping the Ledis server"
        return self.execute_command('PING')

    def select(self, db):
        """Select a Ledis db, ``db`` is integer type"""
        try:
            db = int(db)
        except ValueError:
            db = 0
        return self.execute_command('SELECT', db)


    #### BASIC KEY COMMANDS ####
    def decr(self, name, amount=1):
        """
        Decrements the value of ``key`` by ``amount``.  If no key exists,
        the value will be initialized as 0 - ``amount``
        """
        return self.execute_command('DECRBY', name, amount)

    def decrby(self, name, amount=1):
        """
        Decrements the value of ``key`` by ``amount``.  If no key exists,
        the value will be initialized as 0 - ``amount``
        """
        return self.decr(name, amount)

    def delete(self, *names):
        "Delete one or more keys specified by ``names``"
        return self.execute_command('DEL', *names)

    def exists(self, name):
        "Returns a boolean indicating whether key ``name`` exists"
        return self.execute_command('EXISTS', name)

    def expire(self, name, time):
        """
        Set an expire flag on key ``name`` for ``time`` seconds. ``time``
        can be represented by an integer or a Python timedelta object.
        """
        if isinstance(time, datetime.timedelta):
            time = time.seconds + time.days * 24 * 3600
        return self.execute_command('EXPIRE', name, time)

    def expireat(self, name, when):
        """
        Set an expire flag on key ``name``. ``when`` can be represented
        as an integer indicating unix time or a Python datetime object.
        """
        if isinstance(when, datetime.datetime):
            when = int(mod_time.mktime(when.timetuple()))
        return self.execute_command('EXPIREAT', name, when)

    def get(self, name):
        """
        Return the value at key ``name``, or None if the key doesn't exist
        """
        return self.execute_command('GET', name)

    def __getitem__(self, name):
        """
        Return the value at key ``name``, raises a KeyError if the key
        doesn't exist.
        """
        value = self.get(name)
        if value:
            return value
        raise KeyError(name)

    def getset(self, name, value):
        """
        Set the value at key ``name`` to ``value`` if key doesn't exist
        Return the value at key ``name`` atomically
        """
        return self.execute_command('GETSET', name, value)

    def incr(self, name, amount=1):
        """
        Increments the value of ``key`` by ``amount``.  If no key exists,
        the value will be initialized as ``amount``
        """
        return self.execute_command('INCRBY', name, amount)

    def incrby(self, name, amount=1):
        """
        Increments the value of ``key`` by ``amount``.  If no key exists,
        the value will be initialized as ``amount``
        """

        # An alias for ``incr()``, because it is already implemented
        # as INCRBY ledis command.
        return self.incr(name, amount)

    def mget(self, keys, *args):
        """
        Returns a list of values ordered identically to ``keys``
        """
        args = list_or_args(keys, args)
        return self.execute_command('MGET', *args)

    def mset(self, *args, **kwargs):
        """
        Sets key/values based on a mapping. Mapping can be supplied as a single
        dictionary argument or as kwargs.
        """
        if args:
            if len(args) != 1 or not isinstance(args[0], dict):
                raise LedisError('MSET requires **kwargs or a single dict arg')
            kwargs.update(args[0])
        items = []
        for pair in iteritems(kwargs):
            items.extend(pair)
        return self.execute_command('MSET', *items)

    def set(self, name, value):
        """
        Set the value of key ``name`` to ``value``.
        """
        pieces = [name, value]
        return self.execute_command('SET', *pieces)

    def setnx(self, name, value):
        "Set the value of key ``name`` to ``value`` if key doesn't exist"
        return self.execute_command('SETNX', name, value)

    def ttl(self, name):
        "Returns the number of seconds until the key ``name`` will expire"
        return self.execute_command('TTL', name)

    def persist(self, name):
        "Removes an expiration on name"
        return self.execute_command('PERSIST', name)


    #### LIST COMMANDS ####
    def lindex(self, name, index):
        """
        Return the item from list ``name`` at position ``index``

        Negative indexes are supported and will return an item at the
        end of the list
        """
        return self.execute_command('LINDEX', name, index)

    def llen(self, name):
        "Return the length of the list ``name``"
        return self.execute_command('LLEN', name)

    def lpop(self, name):
        "Remove and return the first item of the list ``name``"
        return self.execute_command('LPOP', name)

    def lpush(self, name, *values):
        "Push ``values`` onto the head of the list ``name``"
        return self.execute_command('LPUSH', name, *values)

    def lrange(self, name, start, end):
        """
        Return a slice of the list ``name`` between
        position ``start`` and ``end``

        ``start`` and ``end`` can be negative numbers just like
        Python slicing notation
        """
        return self.execute_command('LRANGE', name, start, end)

    def rpop(self, name):
        "Remove and return the last item of the list ``name``"
        return self.execute_command('RPOP', name)

    def rpush(self, name, *values):
        "Push ``values`` onto the tail of the list ``name``"
        return self.execute_command('RPUSH', name, *values)

    # SPECIAL COMMANDS SUPPORTED BY LEDISDB
    def lclear(self, name):
        "Delete the key of ``name``"
        return self.execute_command("LCLEAR", name)

    def lmclear(self, *names):
        "Delete multiple keys of ``name``"
        return self.execute_command('LMCLEAR', *names)

    def lexpire(self, name, time):
        """
        Set an expire flag on key ``name`` for ``time`` seconds. ``time``
        can be represented by an integer or a Python timedelta object.
        """
        if isinstance(time, datetime.timedelta):
            time = time.seconds + time.days * 24 * 3600
        return self.execute_command("LEXPIRE", name, time)

    def lexpireat(self, name, when):
        """
        Set an expire flag on key ``name``. ``when`` can be represented as an integer
        indicating unix time or a Python datetime object.
        """
        if isinstance(when, datetime.datetime):
            when = int(mod_time.mktime(when.timetuple()))
        return self.execute_command('LEXPIREAT', name, when)

    def lttl(self, name):
        "Returns the number of seconds until the key ``name`` will expire"
        return self.execute_command('LTTL', name)

    def lpersist(self, name):
        "Removes an expiration on ``name``"
        return self.execute_command('LPERSIST', name)


    #### SORTED SET COMMANDS ####
    def zadd(self, name, *args, **kwargs):
        """
        Set any number of score, element-name pairs to the key ``name``. Pairs
        can be specified in two ways:

        As *args, in the form of: score1, name1, score2, name2, ...
        or as **kwargs, in the form of: name1=score1, name2=score2, ...

        The following example would add four values to the 'my-key' key:
        ledis.zadd('my-key', 1.1, 'name1', 2.2, 'name2', name3=3.3, name4=4.4)
        """
        pieces = []
        if args:
            if len(args) % 2 != 0:
                raise LedisError("ZADD requires an equal number of "
                                 "values and scores")
            pieces.extend(args)
        for pair in iteritems(kwargs):
            pieces.append(pair[1])
            pieces.append(pair[0])
        return self.execute_command('ZADD', name, *pieces)

    def zcard(self, name):
        "Return the number of elements in the sorted set ``name``"
        return self.execute_command('ZCARD', name)

    def zcount(self, name, min, max):
        """
        Return the number of elements in the sorted set at key ``name`` with a score 
        between ``min`` and ``max``. 
        The min and max arguments have the same semantic as described for ZRANGEBYSCORE.
        """
        return self.execute_command('ZCOUNT', name, min, max)

    def zincrby(self, name, value, amount=1):
        "Increment the score of ``value`` in sorted set ``name`` by ``amount``"
        return self.execute_command('ZINCRBY', name, amount, value)

    def zrange(self, name, start, end, desc=False, withscores=False):
        """
        Return a range of values from sorted set ``name`` between
        ``start`` and ``end`` sorted in ascending order.

        ``start`` and ``end`` can be negative, indicating the end of the range.

        ``desc`` a boolean indicating whether to sort the results descendingly

        ``withscores`` indicates to return the scores along with the values.
        The return type is a list of (value, score) pairs

        ``score_cast_func`` a callable used to cast the score return value
        """
        if desc:
            return self.zrevrange(name, start, end, withscores,
                                  score_cast_func)
        pieces = ['ZRANGE', name, start, end]
        if withscores:
            pieces.append('withscores')
        options = {
            'withscores': withscores, 'score_cast_func': int}
        return self.execute_command(*pieces, **options)

    def zrangebyscore(self, name, min, max, start=None, num=None,
                      withscores=False):
        """
        Return a range of values from the sorted set ``name`` with scores
        between ``min`` and ``max``.

        If ``start`` and ``num`` are specified, then return a slice
        of the range.

        ``withscores`` indicates to return the scores along with the values.
        The return type is a list of (value, score) pairs

        `score_cast_func`` a callable used to cast the score return value
        """
        if (start is not None and num is None) or \
                (num is not None and start is None):
            raise LedisError("``start`` and ``num`` must both be specified")
        pieces = ['ZRANGEBYSCORE', name, min, max]
        if start is not None and num is not None:
            pieces.extend(['LIMIT', start, num])
        if withscores:
            pieces.append('withscores')
        options = {
            'withscores': withscores, 'score_cast_func': int}
        return self.execute_command(*pieces, **options)

    def zrank(self, name, value):
        """
        Returns a 0-based value indicating the rank of ``value`` in sorted set
        ``name``
        """
        return self.execute_command('ZRANK', name, value)

    def zrem(self, name, *values):
        "Remove member ``values`` from sorted set ``name``"
        return self.execute_command('ZREM', name, *values)

    def zremrangebyrank(self, name, min, max):
        """
        Remove all elements in the sorted set ``name`` with ranks between
        ``min`` and ``max``. Values are 0-based, ordered from smallest score
        to largest. Values can be negative indicating the highest scores.
        Returns the number of elements removed
        """
        return self.execute_command('ZREMRANGEBYRANK', name, min, max)

    def zremrangebyscore(self, name, min, max):
        """
        Remove all elements in the sorted set ``name`` with scores
        between ``min`` and ``max``. Returns the number of elements removed.
        """
        return self.execute_command('ZREMRANGEBYSCORE', name, min, max)

    def zrevrange(self, name, start, num, withscores=False):
        """
        Return a range of values from sorted set ``name`` between
        ``start`` and ``num`` sorted in descending order.

        ``start`` and ``num`` can be negative, indicating the end of the range.

        ``withscores`` indicates to return the scores along with the values
        The return type is a list of (value, score) pairs

        ``score_cast_func`` a callable used to cast the score return value
        """
        pieces = ['ZREVRANGE', name, start, num]
        if withscores:
            pieces.append('withscores')
        options = {
            'withscores': withscores, 'score_cast_func': int}
        return self.execute_command(*pieces, **options)

    def zrevrangebyscore(self, name, min, max, start=None, num=None,
                         withscores=False):
        """
        Return a range of values from the sorted set ``name`` with scores
        between ``min`` and ``max`` in descending order.

        If ``start`` and ``num`` are specified, then return a slice
        of the range.

        ``withscores`` indicates to return the scores along with the values.
        The return type is a list of (value, score) pairs

        ``score_cast_func`` a callable used to cast the score return value
        """
        if (start is not None and num is None) or \
                (num is not None and start is None):
            raise LedisError("``start`` and ``num`` must both be specified")
        pieces = ['ZREVRANGEBYSCORE', name, min, max]
        if start is not None and num is not None:
            pieces.extend(['LIMIT', start, num])
        if withscores:
            pieces.append('withscores')
        options = {
            'withscores': withscores, 'score_cast_func': int}
        return self.execute_command(*pieces, **options)

    def zrevrank(self, name, value):
        """
        Returns a 0-based value indicating the descending rank of
        ``value`` in sorted set ``name``
        """
        return self.execute_command('ZREVRANK', name, value)

    def zscore(self, name, value):
        "Return the score of element ``value`` in sorted set ``name``"
        return self.execute_command('ZSCORE', name, value)

    # SPECIAL COMMANDS SUPPORTED BY LEDISDB
    def zclear(self, name):
        "Delete key of ``name`` from sorted set"
        return self.execute_command('ZCLEAR', name)

    def zmclear(self, *names):
        "Delete multiple keys of ``names`` from sorted set"
        return self.execute_command('ZMCLEAR', *names)

    def zexpire(self, name, time):
        "Set timeout on key ``name`` with ``time``"
        if isinstance(time, datetime.timedelta):
            time = time.seconds + time.days * 24 * 3600
        return self.execute_command('ZEXPIRE', name, time)

    def zexpireat(self, name, when):
        """
        Set an expire flag on key name for time seconds. time can be represented by
         an integer or a Python timedelta object.
        """

        if isinstance(when, datetime.datetime):
            when = int(mod_time.mktime(when.timetuple()))
        return self.execute_command('ZEXPIREAT', name, when)

    def zttl(self, name):
        "Returns the number of seconds until the key name will expire"
        return self.execute_command('ZTTL', name)

    def zpersist(self, name):
        "Removes an expiration on name"
        return self.execute_command('ZPERSIST', name)



    #### HASH COMMANDS ####
    def hdel(self, name, *keys):
        "Delete ``keys`` from hash ``name``"
        return self.execute_command('HDEL', name, *keys)

    def hexists(self, name, key):
        "Returns a boolean indicating if ``key`` exists within hash ``name``"
        return self.execute_command('HEXISTS', name, key)

    def hget(self, name, key):
        "Return the value of ``key`` within the hash ``name``"
        return self.execute_command('HGET', name, key)

    def hgetall(self, name):
        "Return a Python dict of the hash's name/value pairs"
        return self.execute_command('HGETALL', name)

    def hincrby(self, name, key, amount=1):
        "Increment the value of ``key`` in hash ``name`` by ``amount``"
        return self.execute_command('HINCRBY', name, key, amount)

    def hkeys(self, name):
        "Return the list of keys within hash ``name``"
        return self.execute_command('HKEYS', name)

    def hlen(self, name):
        "Return the number of elements in hash ``name``"
        return self.execute_command('HLEN', name)

    def hmget(self, name, keys, *args):
        "Returns a list of values ordered identically to ``keys``"
        args = list_or_args(keys, args)
        return self.execute_command('HMGET', name, *args)

    def hmset(self, name, mapping):
        """
        Sets each key in the ``mapping`` dict to its corresponding value
        in the hash ``name``
        """
        if not mapping:
            raise DataError("'hmset' with 'mapping' of length 0")
        items = []
        for pair in iteritems(mapping):
            items.extend(pair)
        return self.execute_command('HMSET', name, *items)

    def hset(self, name, key, value):
        """
        Set ``key`` to ``value`` within hash ``name``
        Returns 1 if HSET created a new field, otherwise 0
        """
        return self.execute_command('HSET', name, key, value)

    def hvals(self, name):
        "Return the list of values within hash ``name``"
        return self.execute_command('HVALS', name)

    # SPECIAL COMMANDS SUPPORTED BY LEDISDB
    def hclear(self, name):
        "Delete key ``name`` from hash"
        return self.execute_command('HCLEAR', name)

    def hmclear(self, *names):
        "Delete multiple keys ``names`` from hash"
        return self.execute_command('HMCLEAR', *names)

    def hexpire(self, name, time):
        """
        Set an expire flag on key name for time milliseconds. 
        time can be represented by an integer or a Python timedelta object.
        """
        if isinstance(time, datetime.timedelta):
            time = time.seconds + time.days * 24 * 3600
        return self.execute_command('HEXPIRE', name, time)

    def hexpireat(self, name, when):
        """
        Set an expire flag on key name. when can be represented as an integer representing 
        unix time in milliseconds (unix time * 1000) or a Python datetime object.
        """
        if isinstance(when, datetime.datetime):
            when = int(mod_time.mktime(when.timetuple()))
        return self.execute_command('HEXPIREAT', name, when)

    def httl(self, name):
        "Returns the number of seconds until the key name will expire"
        return self.execute_command('HTTL', name)

    def hpersist(self, name):
        "Removes an expiration on name"
        return self.execute_command('HPERSIST', name)

