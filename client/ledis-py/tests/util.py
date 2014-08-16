#coding: utf-8
import datetime
import time


def current_time():
    return datetime.datetime.now()


def expire_at(minute=1):
    expire_at = current_time() + datetime.timedelta(minutes=minute)
    return expire_at


def expire_at_seconds(minute=1):
    return int(time.mktime(expire_at(minute=minute).timetuple()))

if __name__ == "__main__":
    print expire_at()
    print expire_at_seconds()