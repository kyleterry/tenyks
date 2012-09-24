""" Core of tenyks. Contains Robot, Connection, and IRC/Redis Lines"""
import hashlib
import json
import os
from os.path import join
import re
import sys
import time

import logging
logger = logging.getLogger()

import gevent
from gevent import socket
from gevent import queue
import gevent.monkey
import redis

from constants import PING, PONG
from client import CLIENT_SERVICE_STATUS_ONLINE, CLIENT_SERVICE_STATUS_OFFLINE
import config

gevent.monkey.patch_all()


class Connection(object):

    def __init__(self, name, connection_config):
        self.name = name
        self.connection_config = connection_config
        self.greenlets = []
        self.socket = socket.socket()
        self.socket_connected = False
        self.server_disconnect = False
        self.user_disconnect = False
        self.input_queue = queue.Queue()
        self.input_buffer = ''
        self.output_queue = queue.Queue()
        self.output_buffer = ''
        log_directory = getattr(config, 'LOG_DIR', config.WORKING_DIR)
        self.log_file = join(log_directory, 'irc-%s.log' % self.name)

    def connect(self, reconnecting=False):
        while True:
            try:
                if reconnecting:
                    logger.info('%s: reconnecting...' % self.name)
                else:
                    logger.info('%s: connecting...' % self.name)
                self.socket.connect((self.connection_config['host'],
                    self.connection_config['port']))
                self.socket_connected = True
                self.server_disconnect = False
                logger.info('%s: successfully connected' % self.name)
                break
            except socket.error as e:
                logger.warning('%s: could not connect: retrying...' % self.name)
                time.sleep(5)
        self.spawn_send_and_recv_loops()

    def reconnect(self):
        for greenlet in self.greenlets:
            greenlet.kill()
        self.socket.close()
        self.socket = socket.socket()
        self.connect(reconnecting=True)

    def close(self):
        self.socket.close()
        self.user_disconnect = True

    def spawn_send_and_recv_loops(self):
        self.greenlets.append(gevent.spawn(self.recv_loop))
        self.greenlets.append(gevent.spawn(self.send_loop))

    def recv_loop(self):
        while True:
            data = self.socket.recv(1024).decode('utf-8')
            if not data:
                logger.info('%s: disconnected' % self.name)
                self.socket_connected = False
                self.server_disconnect = True
                break
            with open(self.log_file, 'a+') as log_file:
                log_file.write('receiving: %s' % data)
            self.input_buffer += data
            while '\r\n' in self.input_buffer:
                line, self.input_buffer = self.input_buffer.split('\r\n', 1)
                self.input_queue.put(line)

    def send_loop(self):
        while True:
            if not self.socket_connected:
                break
            try:
                line = self.output_queue.get(timeout=5)
            except queue.Empty:
                continue
            self.output_buffer += line.encode('utf-8', 'replace') + '\r\n'
            while self.output_buffer:
                sent = self.socket.send(self.output_buffer)
                self.output_buffer = self.output_buffer[sent:]
                time.sleep(.5)

    def needs_reconnect(self):
        return not self.socket_connected and self.server_disconnect


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
            prefix, self.command, params = self.irc_prefix_rem(raw_line).groups()
        else:
            prefix, self.command, params = self.irc_noprefix_rem(raw_line).groups()
        self.nick_from, self.user, self.host = self.irc_netmask_rem(prefix).groups()
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
                connection.connectino_config['nick'])
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
        return '<%s: %s>' % (self.nick_from, self.message)

    def to_redis_line(self):
        return redisline_factory({
            'version': 1,
            'type': 'privmsg',
            'data': {
                'nick_from': self.nick_from,
                'message': self.message,
                'irc_channel': self.channel,
                'is_direct': self.direct,
            }
        })


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
                'Instances of this class must be created with version %d of the API' % 
                (self.exact_version))
        self.version = raw_line['version']
        self.type = raw_line['type']
        self.message = raw_line['data']['message']
        self.service_name = raw_line['service_name'] if 'service_name' \
            in raw_line else None
        self.connection_name = connection_name
        self.data = raw_line['data']

    def to_publish(self):
        return json.dumps({
            'version': self.version,
            'type': self.type,
            'service_name': self.service_name,
            'connection_name': self.connection_name,
            'data': self.data,
        })


def redisline_factory(raw_line, version=1):
    if version == 1:
        return RedisLineV1(raw_line)
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
        
        self.prepare_environment()

        self.connections = {}
        self.bootstrap_connections()

    def prepare_environment(self):
        try:
            os.mkdir(config.WORKING_DIR)
            os.mkdir(config.DATA_WORKING_DIR)
        except OSError:
            pass

    def bootstrap_connections(self):
        for name, connection in config.CONNECTIONS.iteritems():
            conn = Connection(name, connection)
            conn.connect()
            self.connections[name] = conn
            self.set_nick_and_join(conn)

    def set_nick_and_join(self, connection):
        self.send(connection.name, "NICK %s" % connection.connection_config['nick'])
        self.send(connection.name, "USER %s %s bla :%s" % (
            connection.connection_config['ident'], connection.connection_config['host'],
            connection.connection_config['realname']))

        # join channels
        for channel in connection.connection_config['channels']:
            self.join(channel, connection)

    def join(self, channel, connection, message=None):
        """ join a irc channel
        """
        password = ''
        if ',' in channel:
            channel, password = channel.split(',')
        chan = '%s %s' % (channel, password)
        self.send(connection.name, 'JOIN %s' % chan.strip())

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

    def pub_sub_loops(self):
        r = redis.Redis(**config.REDIS_CONNECTION)
        pubsub = r.pubsub()
        broadcast_channel = getattr(config, 'BROADCAST_TO_ROBOT_CHANNEL',
            'tenyks.robot.broadcast_to')
        pubsub.subscribe(broadcast_channel)
        for raw_redis_message in pubsub.listen():
            try:
                if raw_redis_message['data'] != 1L:
                    message = json.loads(raw_redis_message['data'])
                    assert False, 'TODO'
                    if message['version'] == 1:
                        if message['type'] == 'privmsg':
                            self.say(message['data']['message'],
                                    channels=message['data']['to'])
                        elif message['type'] == 'register_request':
                            logger.info('Received registration request from %s' % (message['data']['name']))
                            self._handle_service_register_request(message)
                        elif message['type'] == 'unregister_request':
                            logger.info('Received unregistration request from %s' % (message['data']['name']))
                            self._handle_service_unregister_request(message)
            except ValueError:
                logger.info('Pubsub loop: invalid JSON. Ignoring message.')

    def ping_services(self):
        for service in self.fetch_service_info():
            gevent.spawn(self.ping_service, service)
        gevent.spawn_later(30, self.ping_services)

    def ping_service(self, service):
        try:
            service = json.loads(self.r.get('services.%s' % service))
            if service['status'] in (CLIENT_SERVICE_STATUS_ONLINE,
                    CLIENT_SERVICE_STATUS_OFFLINE):
                channel = 'services.%s.ping_response' % service['name']
                to_publish = json.dumps({
                    'version': 1,
                    'type': PING,
                    'data': {
                        'name': service['name'],
                        'message': PING,
                        'channel': channel
                    }
                })
                self.publish(to_publish)
                logger.debug('Robot: waiting for PONG on %s' % channel)
                pong = self.r.blpop(channel, 10) # block for 10 second timeout
                if not pong:
                    logger.debug('Robot: failed to get PONG from %s' % service['name'])
                    logger.info('Robot: putting %s into offline mode' % service['name'])
                    service['status'] = CLIENT_SERVICE_STATUS_OFFLINE
                elif pong == PONG:
                    logger.debug('got PONG from %s' % service['name'])
                    service['status'] = CLIENT_SERVICE_STATUS_ONLINE
                service['last_check'] = time.time()
                to_set = json.dumps(service)
                self.r.set('services.%s' % service['name'], to_set)
                return service
        except Exception, e:
            logger.debug(e)

    def fetch_service_info(self):
        r = redis.Redis(**config.REDIS_CONNECTION)
        services_set_key = getattr(config, 'SERVICES_SET_KEY',
                'tenyks.services')
        return r.smembers(services_set_key)

    def connection_worker(self, connection):
        while True:
            if connection.user_disconnect:
                break
            if connection.needs_reconnect():
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
                        'QUIT :%s' % getattr(self, 'exit_message', ''))
                connection.close()
            sys.exit('Bye.')


def make_and_run_robot():
    robot = Robot()
    robot.run()


if __name__ == '__main__':
    make_and_run_robot()
