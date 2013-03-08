from redis import Redis

import tenyks.config as config

def pubsub_factory(channel):
    """
    Returns a Redis.pubsub object subscribed to `channel`.
    """
    rdb = Redis(**config.REDIS_CONNECTION)
    pubsub = rdb.pubsub()
    pubsub.subscribe(channel)
    return pubsub
