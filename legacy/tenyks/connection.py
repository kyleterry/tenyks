from datetime import datetime
import os
import time

import logging

import gevent
from gevent import socket
from gevent import queue
from gevent import ssl


def get_certs_bundle():
    return os.path.join(os.path.dirname(__file__), 'cacert.pem')


class Connection(object):

    def __init__(self, name, **config):
        self.name = name
        self.config = config
        self.connect_attempts = 0
        self.using_ssl = ('ssl' in config and config['ssl'])
        self.greenlets = []
        self.socket = self._fetch_socket()
        self.socket_connected = False
        self.server_disconnect = False
        self.user_disconnect = False
        self.input_queue = queue.Queue()
        self.input_buffer = ''
        self.output_queue = queue.Queue()
        self.output_buffer = ''
        self.logger = logging.getLogger('tenyks.connection.' + self.name)

    def _fetch_socket(self):
        if self.using_ssl:
            if 'ssl_version' in self.config and self.config['ssl_version']:
                if self.config['ssl_version'] == 2:
                    ssl_version = ssl.PROTOCOL_SSLv2
                elif self.config['ssl_version'] == 3:
                    ssl_version = ssl.PROTOCOL_SSLv3
                else:
                    ssl_version = ssl.PROTOCOL_SSLv23
            else:
                # SSLv23 is the default for python ssl and gevent ssl
                ssl_version = ssl.PROTOCOL_SSLv23
            return ssl.wrap_socket(
                socket.socket(),
                cert_reqs=ssl.CERT_REQUIRED,
                ca_certs=get_certs_bundle(),
                ssl_version=ssl_version)
        else:
            return socket.socket()

    def connect(self, reconnecting=False):
        while True:
            self.connect_attempts += 1
            max_retries = self.config.get('retries')
            if max_retries and (self.connect_attempts > max_retries):
                raise Exception("retries exceeded")
            try:
                if reconnecting:
                    self.logger.info('Reconnecting...')
                else:
                    self.logger.info('Connecting...')
                self.socket.connect((self.config['host'],
                    self.config['port']))
                self.socket_connected = True
                self.server_disconnect = False
                self.last_ping = datetime.now()
                self.logger.info('Successfully connected')
                break
            except socket.error as e:
                self.logger.warning('Could not connect: retrying...')
                self.logger.debug(unicode(e, errors='replace'))
                time.sleep(5)

    def post_connect(self):
        self.connect_attempts = 0
        self.spawn_send_and_recv_loops()

    def reconnect(self):
        for greenlet in self.greenlets:
            greenlet.kill()
        self.socket.close()
        self.socket = self._fetch_socket()
        self.connect(reconnecting=True)

    @property
    def needs_reconnect(self):
        return not self.socket_connected and self.server_disconnect

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
                self.logger.info('disconnected')
                self.socket_connected = False
                self.server_disconnect = True
                break
            self.logger.debug(u'<- IRC: {data}'.format(data=data))
            self.input_buffer += data
            while '\r\n' in self.input_buffer:
                line, self.input_buffer = self.input_buffer.split('\r\n', 1)
                self.input_queue.put(line)

    def send(self, message):
        """
        send a message to an IRC connection
        """
        message = str(message).strip()
        self.output_queue.put(message)
        self.logger.info(u'Connection -> {connection}: {message}'.format(
            connection=self, message=message))

    def send_loop(self):
        while True:
            line = self.output_queue.get()
            self.output_buffer += line.encode('utf-8', 'replace') + '\r\n'
            while self.output_buffer:
                sent = self.socket.send(self.output_buffer)
                self.logger.info(u'-> IRC: {data}'.format(
                    data=self.output_buffer))
                self.output_buffer = self.output_buffer[sent:]
                time.sleep(.5)
