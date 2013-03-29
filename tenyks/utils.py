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


def parse_irc_message(s):
    prefix = ''
    trailing = []
    if not s:
        return None
    if s[0] == ':':
        prefix, s = s[1:].split(' ', 1)
    if s.find(' :') != -1:
        s, trailing = s.split(' :', 1)
        args = s.split()
        args.append(trailing)
    else:
        args = s.split()
    command = args.pop(0)
    return prefix, command, args


irc_prefix_re = re.compile(r':?([^!@]*)!?([^@]*)@?(.*)').match
def parse_irc_prefix(prefix):
     nick, user, host = irc_prefix_re(prefix).groups()
     return nick, user, host


def parse_args(args):
    target, message = args
    return target, message


def get_privmsg_data(connection, prefix, command, args):
    data = {}
    if command == 'PRIVMSG':
        nick, user, host = parse_irc_prefix(prefix)
        target, message = parse_args(args)
        data['nick'] = nick
        data['user'] = user
        data['host'] = host
        data['mask'] = '{user}@{host}'.format(user=user, host=host)
        data['command'] = command
        data['connection'] = connection.name
        data['target'] = target
        if target.startswith('#'):
            data['from_channel'] = True
        else:
            data['from_channel'] = False
        data['full_message'] = message
        data['direct'] = (
                message.startswith(connection.config['nick']) or
                not data['from_channel'])
        if data['direct'] and data['from_channel']:
            data['payload'] = ' '.join(message.split()[1:])
        else:
            data['payload'] = data['full_message']
    return data
