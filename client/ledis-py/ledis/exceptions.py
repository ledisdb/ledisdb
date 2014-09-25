"Core exceptions raised by the LedisDB client"


class LedisError(Exception):
    pass

class ServerError(LedisError):
    pass


class ConnectionError(ServerError):
    pass


class BusyLoadingError(ConnectionError):
    pass


class TimeoutError(LedisError):
    pass


class InvalidResponse(ServerError):
    pass


class ResponseError(LedisError):
    pass


class DataError(LedisError):
    pass


class ExecAbortError(ResponseError):
    pass

class TxNotBeginError(LedisError):
    pass