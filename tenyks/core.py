""" Core of tenyks. Contains Robot and IRC/Redis Lines"""
import gevent.monkey
gevent.monkey.patch_all()

from datetime import datetime
import hashlib
import json
import os
import re
import sys
import time

import logging
logger = logging.getLogger('tenyks')

import gevent
from gevent import queue
import redis

from tenyks.config import settings, collect_settings
from tenyks.connection import Connection
from tenyks.utils import pubsub_factory, parse_irc_message, get_privmsg_data
from tenyks.middleware import CORE_MIDDLEWARE


if hasattr(settings, 'MIDDLEWARE'):
    CORE_MIDDLEWARE += settings.MIDDLEWARE


class Robot(object):
    """
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
        # TODO FIX THIS MESS
        try:
            os.mkdir(settings.WORKING_DIR)
        except OSError:
            # Already exists
            pass
        try:
            os.mkdir(settings.DATA_WORKING_DIR)
        except OSError:
            # Already exists
            pass

    def bootstrap_connections(self):
        for name, connection in settings.CONNECTIONS.iteritems():
            conn = Connection(name, **connection)
            conn.connect()
            self.connections[name] = conn
            self.set_nick_and_join(conn)

    def set_nick_and_join(self, connection):
        self.send(connection.name, 'NICK {nick}'.format(
            nick=connection.config['nick']))
        self.send(connection.name, 'USER {ident} {host} bla :{realname}'.format(
            ident=connection.config['ident'],
            host=connection.config['host'],
            realname=connection.config['realname']))

        # join channels
        for channel in connection.config['channels']:
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
        connection.last_ping = datetime.now()
        logger.debug(
            '{connection} Connection Worker: last_ping: {dt}'.format(
                connection=connection.name, dt=connection.last_ping))
        message = message.replace('PING', 'PONG')
        self.send(connection.name, message)

    def send(self, connection, message):
        """
        send a message to an IRC connection
        """
        message = message.strip()
        self.connections[connection].output_queue.put(message)
        logger.info('Robot -> {connection}: {message}'.format(
            connection=connection, message=message))

    def say(self, connection, message, channels=[]):
        for channel in channels:
            message = 'PRIVMSG {channel} :{message}\r\n'.format(
                    channel=channel, message=message)
            self.send(connection, message)

    def broadcast_loop(self):
        """
        Pop things off the broadcast_queue and create jobs for them.
        """
        while True:
            data = self.broadcast_queue.get()
            gevent.spawn(self.broadcast_worker, data)

    def broadcast_worker(self, data):
        """
        This worker will broadcast a message to the service broadcast channel
        """
        r = redis.Redis(**settings.REDIS_CONNECTION)
        broadcast_channel = getattr(settings, 'BROADCAST_TO_SERVICES_CHANNEL',
            'tenyks.services.broadcast_to')
        r.publish(broadcast_channel, json.dumps(data))

    def handle_incoming_redis_messages(self):
        """
        {
            'payload': 'this is a message to IRC',
            'target': '#test',
            'command': 'PRIVMSG',
            'connection': 'freenode',
        }
        """
        broadcast_channel = getattr(settings, 'BROADCAST_TO_ROBOT_CHANNEL',
            'tenyks.robot.broadcast_to')
        pubsub = pubsub_factory(broadcast_channel)
        for raw_redis_message in pubsub.listen():
            logger.info('Robot <- {data}'.format(
                data=json.dumps(raw_redis_message)))
            try:
                if raw_redis_message['data'] != 1L:
                    message = json.loads(raw_redis_message['data'])
                    if message['command'].lower() == 'privmsg':
                        self.say(message['connection'],
                                    message['payload'],
                                    channels=[message['target']])
            except ValueError:
                logger.info('Robot Pubsub: invalid JSON. Ignoring message.')

    def connection_worker(self, connection):
        while True:
            if connection.user_disconnect:
                break
            needs_reconnect = connection.needs_reconnect
            ping_delta = datetime.now() - connection.last_ping
            no_ping = ping_delta.seconds > 5 * 60
            if needs_reconnect or no_ping:
                connection.reconnect()
                self.set_nick_and_join(connection)
            try:
                raw_line = connection.input_queue.get(timeout=5)
            except queue.Empty:
                continue
            if raw_line.startswith('PING'):
                self.handle_irc_ping(connection, raw_line)
                continue
            else:
                data = get_privmsg_data(
                    connection, *parse_irc_message(raw_line))
                if data:
                    for middleware in CORE_MIDDLEWARE:
                        data = middleware(connection, data)
                    self.broadcast_queue.put(data)
        logger.info('{connection} Connection Worker: worker shutdown'.format(
            connection=connection.name))

    def run(self):
        try:
            while True:
                self.workers = []
                for name, connection in self.connections.iteritems():
                    self.workers.append(
                            gevent.spawn(self.connection_worker, connection))
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
                            message=getattr(self, 'exit_message', 'I\'m out!')))
                connection.close()
            sys.exit('Bye.')


def main():
    collect_settings()
    robot = Robot()
    robot.run()


if __name__ == '__main__':
    main()
