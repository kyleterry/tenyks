import re

from redis import Redis

from tenyks.config import settings

def pubsub_factory(channel):
    """
    Returns a Redis.pubsub object subscribed to `channel`.
    """
    rdb = Redis(**settings.REDIS_CONNECTION)
    pubsub = rdb.pubsub()
    pubsub.subscribe(channel)
    return pubsub

irc_prefix_re = re.compile(r':?([^!@]*)!?([^@]*)@?(.*)').match
def parse_irc_prefix(prefix):
     nick, user, host = irc_prefix_re(prefix).groups()
     return {
        'nick': nick, 
        'user': user, 
        'host': host
    }
