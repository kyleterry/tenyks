from datetime import datetime
from os.path import join
import time

import logging

import gevent
from gevent import socket
from gevent import queue
import gevent.monkey

import tenyks.config as config


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
        self.logger = logging.getLogger(self.name)

    def connect(self, reconnecting=False):
        while True:
            try:
                if reconnecting:
                    self.logger.info('Reconnecting...')
                else:
                    self.logger.info('Connecting...')
                self.socket.connect((self.connection_config['host'],
                    self.connection_config['port']))
                self.socket_connected = True
                self.server_disconnect = False
                self.last_ping = datetime.now()
                self.logger.info('Successfully connected')
                break
            except socket.error as e:
                self.logger.warning('Could not connect: retrying...')
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
                self.logger.info('disconnected')
                self.socket_connected = False
                self.server_disconnect = True
                break
            self.logger.info('<- IRC: {data}'.format(data=data))
            with open(self.log_file, 'a+') as log_file:
                log_file.write('receiving: %s' % data)
            self.input_buffer += data
            while '\r\n' in self.input_buffer:
                line, self.input_buffer = self.input_buffer.split('\r\n', 1)
                self.input_queue.put(line)

    def send_loop(self):
        while True:
            line = self.output_queue.get()
            self.output_buffer += line.encode('utf-8', 'replace') + '\r\n'
            while self.output_buffer:
                sent = self.socket.send(self.output_buffer)
                self.logger.info('-> IRC: {data}'.format(
                    data=self.output_buffer))
                self.output_buffer = self.output_buffer[sent:]
                time.sleep(.5)

    def needs_reconnect(self):
        return not self.socket_connected and self.server_disconnect
