""" Core of tenyks. Contains Robot, Connection, and IRC/Redis Lines"""
from datetime import datetime
import hashlib
import json
import os
import re
import sys
import time

import logging
logger = logging.getLogger()

import gevent
from gevent import queue
import gevent.monkey
import redis

from tenyks.client import (CLIENT_SERVICE_STATUS_ONLINE,
        CLIENT_SERVICE_STATUS_OFFLINE)
from tenyks.constants import PING, PONG
import tenyks.config as config
from tenyks.connection import Connection
from tenyks.utils import pubsub_factory

gevent.monkey.patch_all()


class IrcLine(object):
    irc_prefix_rem = re.compile(r'(.*?) (.*?) (.*)').match
    irc_noprefix_rem = re.compile(r'()(.*?) (.*)').match
    irc_netmask_rem = re.compile(r':?([^!@]*)!?([^@]*)@?(.*)').match
    irc_param_ref = re.compile(r'(?:^|(?<= ))(:.*|[^ ]+)').findall

    def __init__(self, connection, raw_line):
        self.line_id = hashlib.sha256(
                str(time.time()) + raw_line).hexdigest()[:10]
        self.raw_line = raw_line
        self.connection = connection
        if raw_line.startswith(":"):  # has a prefix
            prefix, self.command, params = self.irc_prefix_rem(
                    raw_line).groups()
        else:
            prefix, self.command, params = self.irc_noprefix_rem(
                    raw_line).groups()
        self.nick_from, self.user, self.host = self.irc_netmask_rem(
                prefix).groups()
        self.mask = self.user + "@" + self.host
        self.paramlist = self.irc_param_ref(params)
        lastparam = ""
        if self.paramlist:
            if self.paramlist[-1].startswith(':'):
                self.paramlist[-1] = self.paramlist[-1][1:]
            lastparam = self.paramlist[-1]
        self.channel = self.paramlist[0]
        self.message = lastparam.lower()
        self.direct = self.message.startswith(
                connection.connection_config['nick'])
        self.verb = ''
        if self.message:
            try:
                if self.direct:
                    self.verb = self.message.split()[1]
                else:
                    self.verb = self.message.split()[0]
            except IndexError:
                self.verb = 'confused'
        if self.direct:
            # remove 'BOTNICK: ' from message
            self.message = " ".join(self.message.split()[1:])

    def __repr__(self):
        return '<{nick}: {message}>'.format(
                nick=self.nick_from, message=self.message)

    def to_redis_line(self):
        return redisline_factory({
            'version': 1,
            'type': 'privmsg',
            'nick_from': self.nick_from,
            'irc_channel': self.channel,
            'direct': self.direct,
            'payload': self.message,
        }, connection=self.connection)


class RedisLineVersionMismatch(Exception):
    pass


class RedisLine(object):

    def __init__(self):
        raise NotImplementedError


class RedisLineV1(RedisLine):

    exact_version = 1

    def __init__(self, raw_line, connection_name=None):
        if type(raw_line) in (str, unicode,):
            raw_line = json.loads(raw_line)
        if raw_line['version'] != self.exact_version:
            raise RedisLineVersionMismatch(
                'Instances of this class must be created '
                'with version {version} of the API'.format(
                (self.exact_version)))
        self.version = raw_line['version']
        self.message_type = raw_line['type']
        self.payload = raw_line['payload']
        self.irc_channel = raw_line['irc_channel']
        self.direct = raw_line['direct']
        self.nick_from = raw_line['nick_from']
        self.service_name = raw_line['service_name'] if 'service_name' \
            in raw_line else None
        self.connection_name = connection_name

    def to_publish(self):
        return json.dumps({
            'version': self.version,
            'type': self.message_type,
            'service_name': self.service_name,
            'connection_name': self.connection_name,
            'irc_channel': self.irc_channel,
            'direct': self.direct,
            'nick_from': self.nick_from,
            'payload': self.payload,
        })


def redisline_factory(raw_line, connection, version=1):
    if version == 1:
        return RedisLineV1(raw_line, connection_name=connection.name)
    return None


class Robot(object):
    """
    Platforms are IRC, Email, SMS, ect...

    Design Rules:

    * The Robot has access to services to handle commands.
    * Services send and receive data.
    * The Robot's brain is, effectively, the services.
    * The Robot will forever ONLY relay data to and from platforms (IRC, Email)
      to and from services and handle connection to platforms.
    * The Robot will never handle commands sent from platforms or services.
    * _Almost_ everything is evented.
    """

    def __init__(self):
        self.broadcast_queue = queue.Queue()
        gevent.spawn(self.broadcast_loop)
        gevent.spawn(self.handle_incoming_redis_messages)

        self.prepare_environment()

        self.connections = {}
        self.bootstrap_connections()

    def prepare_environment(self):
        try:
            os.mkdir(config.WORKING_DIR)
            os.mkdir(config.DATA_WORKING_DIR)
        except OSError:
            # Already exists
            pass

    def bootstrap_connections(self):
        for name, connection in config.CONNECTIONS.iteritems():
            conn = Connection(name, connection)
            conn.connect()
            self.connections[name] = conn
            self.set_nick_and_join(conn)

    def set_nick_and_join(self, connection):
        self.send(connection.name, 'NICK {nick}'.format(
            nick=connection.connection_config['nick']))
        self.send(connection.name, 'USER {ident} {host} bla :{realname}'.format(
            ident=connection.connection_config['ident'],
            host=connection.connection_config['host'],
            realname=connection.connection_config['realname']))

        # join channels
        for channel in connection.connection_config['channels']:
            self.join(channel, connection)

    def join(self, channel, connection, message=None):
        """ join a irc channel
        """
        password = ''
        if ',' in channel:
            channel, password = channel.split(',')
        chan = '{channel} {password}'.format(
                channel=channel, password=password)
        self.send(connection.name, 'JOIN {channel}'.format(
            channel=chan.strip()))

    def handle_irc_ping(self, connection, message):
        """
        always returns None
        """
        message = message.replace('PING', 'PONG')
        self.send(connection.name, message)

    def send(self, connection, message):
        """
        send a message to an IRC connection
        """
        message = message.strip()
        self.connections[connection].output_queue.put(message)

    def say(self, connection, message, channels=[]):
        for channel in channels:
            message = 'PRIVMSG {channel} :{message}\r\n'.format(
                    channel=channel, message=message)
            logger.info('Sending {connection}: {message}'.format(
                connection=connection, message=message))
            self.send(connection, message)

    def broadcast_loop(self):
        """
        Pop things off the broadcast_queue and create jobs for them.
        """
        while True:
            irc_line = self.broadcast_queue.get()
            gevent.spawn(self.broadcast_worker, irc_line)

    def broadcast_worker(self, irc_line):
        """
        This worker will broadcast a message to the service broadcast channel
        """
        r = redis.Redis(**config.REDIS_CONNECTION)
        broadcast_channel = getattr(config, 'BROADCAST_TO_SERVICES_CHANNEL',
            'tenyks.services.broadcast_to')
        r.publish(broadcast_channel, irc_line.to_redis_line().to_publish())

    def handle_incoming_redis_messages(self):
        broadcast_channel = getattr(config, 'BROADCAST_TO_ROBOT_CHANNEL',
            'tenyks.robot.broadcast_to')
        pubsub = pubsub_factory(broadcast_channel)
        for raw_redis_message in pubsub.listen():
            logger.debug('robot got: {data}'.format(data=json.dumps(raw_redis_message)))
            try:
                if raw_redis_message['data'] != 1L:
                    message = json.loads(raw_redis_message['data'])
                    if message['version'] == 1:
                        if message['type'] == 'privmsg':
                            self.say(message['connection_name'],
                                      message['payload'],
                                      channels=[message['irc_channel']])
            except ValueError:
                logger.info('Pubsub loop: invalid JSON. Ignoring message.')

    def connection_worker(self, connection):
        while True:
            if connection.user_disconnect:
                break
            needs_reconnect = connection.needs_reconnect()
            ping_delta = datetime.now() - connection.last_ping
            no_ping = ping_delta.seconds > 4 * 60
            if needs_reconnect or no_ping:
                connection.reconnect()
                self.set_nick_and_join(connection)
            try:
                raw_line = connection.input_queue.get(timeout=5)
            except queue.Empty:
                continue
            if raw_line.startswith('PING'):
                connection.last_ping = datetime.now()
                logger.debug(
                    '{connection} connection_worker: setting last_ping to {dt}'.format(
                        connection=connection.name, dt=connection.last_ping))
                self.handle_irc_ping(connection, raw_line)
                continue
            else:
                irc_line = IrcLine(connection, raw_line)
                self.broadcast_queue.put(irc_line)
        logger.info('Connection Worker: worker shutdown')

    def run(self):
        try:
            while True:
                self.workers = []
                for name, connection in self.connections.iteritems():
                    self.workers.append(gevent.spawn(self.connection_worker, connection))
                gevent.joinall(self.workers)
                break
        except KeyboardInterrupt:
            logger.info('Robot: shutting down: user disconnect')
            for name, connection in self.connections.iteritems():
                connection.close()
        finally:
            for name, connection in self.connections.iteritems():
                self.send(connection.name,
                        'QUIT :{message}'.format(
                            message=getattr(self, 'exit_message', '')))
                connection.close()
            sys.exit('Bye.')


def make_and_run_robot():
    robot = Robot()
    robot.run()


if __name__ == '__main__':
    make_and_run_robot()
