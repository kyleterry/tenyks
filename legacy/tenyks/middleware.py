from datetime import datetime
import re

import logging
logger = logging.getLogger('tenyks')

from tenyks.utils import parse_irc_prefix
from tenyks import commands

def irc_parse(robot, connection, data):
    command_re = r'^(:(?P<prefix>\S+) )?(?P<cmd>\S+)( (?!:)(?P<args>.+?))?( :(?P<trail>.+))?$'
    match_obj = re.match(command_re, data)
    data = {
        'prefix': match_obj.group('prefix'),
        'command': match_obj.group('cmd'),
        'args': match_obj.group('args'),
        'trail': match_obj.group('trail'),
        'raw': data
    }
    return data

def irc_extract(robot, connection, data):
    if data['command'] == 'PRIVMSG':
        data.update(parse_irc_prefix(data['prefix']))
        message = data['trail']
        data['mask'] = '{user}@{host}'.format(user=data['user'],
            host=data['host'])
        data['connection'] = connection.name
        data['target'] = data['args']
        data['from_channel'] = data['target'].startswith('#')
        data['full_message'] = message
        data['direct'] = (
                message.startswith(connection.nick) or
                not data['from_channel'])

        # TODO CLEAN THIS SHIT UP
        if data['direct'] and data['from_channel']:
            data['private_message'] = False
            data['payload'] = ' '.join(message.split()[1:])
        elif data['from_channel']:
            data['private_message'] = False
            data['payload'] = data['full_message']
        else:
            data['private_message'] = True
            data['payload'] = data['full_message']
    return data

def irc_autoreply(robot, connection, data):
    if data['command'] == 'PING':
        connection.last_ping = datetime.now()
        logger.debug(
            '{connection} autoreply Middleware: last_ping: {dt}'.format(
                connection=connection.name, dt=connection.last_ping))
        reply = data['raw'].replace('PING', 'PONG')
        connection.output_queue.put(reply)
    # Authenticated to server
    elif data['command'] == '001':
        if connection.config.get('commands'):
            for command in connection.config['commands']:
                robot.run_command(connection, command)
        for channel in connection.config['channels']:
            password = ''
            if ',' in channel:
                channel, password = channel.split(',')
            connection.send(commands.JOIN(
                channel=channel.strip(),
                password=password.strip()))
    # Nickname in use
    elif data['command'] == '433':
        nicks = connection.config.get('nicks')
        offset = (nicks.index(connection.nick) + 1) % len(nicks)
        connection.nick = connection.config['nicks'][offset]
        connection.send(commands.NICK(nick=connection.nick))
    return data

def admin_middlware(robot, connection, data):
    conf = connection.config
    data['admin'] = data.get('nick') in conf['admins']
    return data

CORE_MIDDLEWARE = (
    irc_parse,
    irc_extract,
    irc_autoreply,
    admin_middlware
)
