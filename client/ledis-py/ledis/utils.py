
def from_url(url, db=None, **kwargs):
    """
    Returns an active Ledis client generated from the given database URL.

    Will attempt to extract the database id from the path url fragment, if
    none is provided.
    """
    from ledis.client import Ledis
    return Ledis.from_url(url, db, **kwargs)
