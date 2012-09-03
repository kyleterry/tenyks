import os
import sys
import logging
import time
from ssl import CERT_NONE
from threading import Thread, Lock

import redis
import gevent
from gevent import monkey
from gevent import socket, queue
from gevent.ssl import wrap_socket

from tenyks.exceptions import ConnectionConfigurationError
from tenyks.logs import (make_log, make_debug_log, INFO, WARNING, ERROR,
        OK, FAILED)
from tenyks import settings


class Connection(object):

    def __init__(self, config):
        self.config = config
        self.greenlets = []
        self._ibuffer = ''
        self._obuffer = ''
        self.input_queue = queue.Queue()
        self.output_queue = queue.Queue()
        self.socket = self.create_socket()

    @classmethod
    def validate_config(self, config):
        #must_have = ['host', 'port', 'nick', 'ident', 'realname', 'channels']
        #for key in config:
        #    try:
        #        must_have.remove(key)
        #    except ValueError:
        #        raise ConnectionConfigurationError('%s is missing from connection configuration' % key)
        return config

    @classmethod
    def from_config(cls, config):
        cls.validate_config(config)
        return cls(config)

    def kill_greenlets(self):
        for greenlet in self.greenlets:
            greenlet.kill()

    def create_socket(self):
        """
            This method exists to create ssl sockets, also
        """
        return socket.socket()

    def close_socket(self):
        self.socket.close()

    def connect(self):
        MAX_TRIES = 5
        ATTEMPTS = 0
        RETRY = 5
        self.kill_greenlets()
        while True:
            try:
                self.socket.connect((self.config['host'], self.config['port']))
                break
            except socket.error as e:
                make_log('%s Connecting to host...' % FAILED, type=ERROR)
                make_log('Retrying in %d' % RETRY, type=INFO)
                time.sleep(RETRY)
                ATTEMPTS += 1
                if ATTEMPTS == MAX_TRIES:
                    sys.exit('Failed to connect to host. Exiting...')
        make_log('%s Connecting to host...' % OK, type=INFO)
        gevent.spawn(self.spawn_send_and_recv_loops)

    def spawn_send_and_recv_loops(self):
        self.greenlets.append(gevent.spawn(self.recv_loop))
        self.greenlets.append(gevent.spawn(self.send_loop))

    def recv_loop(self):
        while True:
            data = self.socket.recv(1024).decode('utf')
            self._ibuffer += data
            while '\r\n' in self._ibuffer:
                line, self._ibuffer = self._ibuffer.split('\r\n', 1)
                self.input_queue.put(line)

    def send_loop(self):
        while True:
            line = self.output_queue.get()
            self._obuffer += line.encode('utf-8', 'replace') + '\r\n'
            while self._obuffer:
                sent = self.socket.send(self._obuffer)
                self._obuffer = self._obuffer[sent:]
                time.sleep(.5)


class Line(object):
    pass


class Tenyks(object):

    def __init__(self):
        self.connections = {}
        self.bootstrap_connections()

    def bootstrap_connections(self):
        for conn_name, conn_config in settings.CONNECTIONS.iteritems():
            connection = Connection.from_config(conn_config)
            connection.connect()
            self.connections[conn_name] = connection
            self.set_nick_and_join(conn_name, conn_config)

    def set_nick_and_join(self, connection_name, conn_config):
        self.send('NICK %s' % conn_config['nick'], connection_name)
        self.send('USER %s %s blah :%s' % (
            conn_config['ident'], conn_config['host'], conn_config['realname']),
            connection_name)
        for channel in conn_config['channels']:
            self.send('JOIN %s' % channel, connection_name)

    def send(self, message, connection_name):
        message = message.strip()
        make_log('Sent: %s' % message)
        self.connections[connection_name].output_queue.put(message)

    def run(self):
        while True:
            pass


if __name__ == '__main__':
    bot = Tenyks()
    bot.run()
