import sys
import time

import gevent
from gevent import socket
from gevent import queue
import gevent.monkey

import settings

gevent.monkey.patch_all()


class Connection(object):

    def __init__(self, connection_name, config):
        self.connection_name = connection_name
        self.config = config
        self.greenlets = []
        self.socket = socket.socket()
        self.input_queue = queue.Queue()
        self.input_buffer = ''
        self.output_queue = queue.Queue()
        self.output_buffer = ''

    def connect(self):
        while True:
            try:
                self.socket.connect((self.config['host'], self.config['port']))
                break
            except socket.error as e:
                print 'Retrying'
        self.spawn_send_and_recv_loops()

    def spawn_send_and_recv_loops(self):
        self.greenlets.append(gevent.spawn(self.recv_loop))
        self.greenlets.append(gevent.spawn(self.send_loop))

    def recv_loop(self):
        while True:
            data = self.socket.recv(1024).decode('utf-8')
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
                self.output_buffer = self.output_buffer[sent:]
                time.sleep(.5)


class Robot(object):

    def __init__(self):
        self.connections = {}
        self.bootstrap_connections()

    def bootstrap_connections(self):
        for name, connection in settings.CONNECTIONS.iteritems():
            conn = Connection(name, connection)
            conn.connect()
            self.connections[name] = conn
            self.set_nick_and_join(conn)

    def set_nick_and_join(self, connection):
        self.send(connection.connection_name, "NICK %s" % connection.config['nick'])
        self.send(connection.connection_name, "USER %s %s bla :%s" % (
            connection.config['ident'], connection.config['host'],
            connection.config['realname']))

        # join channels
        for channel in connection.config['channels']:
            self.join(channel, connection)

    def join(self, channel, connection, message=None):
        """ join a irc channel
        """
        password = ''
        if ',' in channel:
            channel, password = channel.split(',')
        chan = '%s %s' % (channel, password)
        self.send(connection.connection_name, 'JOIN %s' % chan.strip())

    def send(self, connection, message):
        message = message.strip()
        self.connections[connection].output_queue.put(message)

    def run(self):
        try:
            while True:
                for name, connection in self.connections.iteritems():
                    raw_line = connection.input_queue.get()
                    print raw_line
        except KeyboardInterrupt:
            sys.exit('Bye.')


robot = Robot()
robot.run()
