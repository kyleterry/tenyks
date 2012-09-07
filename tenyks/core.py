import re
import sys
import time
import json
from os.path import join

import gevent
from gevent import socket
from gevent import queue
import gevent.monkey

import settings

gevent.monkey.patch_all()


class Connection(object):

    def __init__(self, name, config):
        self.name = name
        self.config = config
        self.greenlets = []
        self.socket = socket.socket()
        self.input_queue = queue.Queue()
        self.input_buffer = ''
        self.output_queue = queue.Queue()
        self.output_buffer = ''
        log_directory = getattr(settings, 'LOG_DIRECTORY', '/tmp/tenyks')
        self.log_file = join(log_directory.split('/'), 'irc-%s.log' % self.name)

    def connect(self):
        while True:
            try:
                self.socket.connect((self.config['host'], self.config['port']))
                break
            except socket.error as e:
                print 'Could not connect to %s: Retrying' % self.name
        self.spawn_send_and_recv_loops()

    def close(self):
        self.socket.close()

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
            print 'sending: %s' % line
            self.output_buffer += line.encode('utf-8', 'replace') + '\r\n'
            while self.output_buffer:
                sent = self.socket.send(self.output_buffer)
                self.output_buffer = self.output_buffer[sent:]
                time.sleep(.5)


class IrcLine(object):
    irc_prefix_rem = re.compile(r'(.*?) (.*?) (.*)').match
    irc_noprefix_rem = re.compile(r'()(.*?) (.*)').match
    irc_netmask_rem = re.compile(r':?([^!@]*)!?([^@]*)@?(.*)').match
    irc_param_ref = re.compile(r'(?:^|(?<= ))(:.*|[^ ]+)').findall

    def __init__(self, connection, raw_line):
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
        self.direct = self.message.startswith(connection.config['nick'])
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


class RedisLineVersionMismatch(Exception):
    pass


class RedisLine(object):

    def __init__(self):
        raise NotImplementedError


class RedisLineV1(RedisLine):

    exact_version = 1

    def __init__(self, raw_line):
        if type(raw_line) in (str, unicode,):
            raw_line = json.loads(raw_line)
        if raw_line.version != self.exact_version:
            raise RedisLineVersionMismatch(
                'Instances of this class must be created with version %d of the API' % 
                (self.exact_version))
        self.version = raw_line.version
        self.type = raw_line.type
        self.message = raw_line.data['message']
        self.service_name = raw_line['service_name']
        self.data = raw_line.data

    def to_publish(self):
        return json.dumps()


def redisline_factory(raw_line):
    return RedisLineV1(raw_line)


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
        self.connections = {}
        self.bootstrap_connections()

    def bootstrap_connections(self):
        for name, connection in settings.CONNECTIONS.iteritems():
            conn = Connection(name, connection)
            conn.connect()
            self.connections[name] = conn
            self.set_nick_and_join(conn)

    def set_nick_and_join(self, connection):
        self.send(connection.name, "NICK %s" % connection.config['nick'])
        self.send(connection.name, "USER %s %s bla :%s" % (
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
        self.send(connection.name, 'JOIN %s' % chan.strip())

    def handle_irc_ping(self, connection, message):
        """
        always returns None
        """
        message = message.replace('PING', 'PONG')
        self.send(connection.name, self.make_pong(message))

    def send(self, connection, message):
        """
        send a message to an IRC connection
        """
        message = message.strip()
        self.connections[connection].output_queue.put(message)

    def broadcast(self, line):
        """
        Broadcase a line from IRC to all the services
        """
        pass

    def run(self):
        try:
            while True:
                for name, connection in self.connections.iteritems():
                    raw_line = connection.input_queue.get()
                    if raw_line.startswith('PING'):
                        self.handle_irc_ping(connection, raw_line)
                        continue
                    irc_line = IrcLine(connection, raw_line)
                    self.broadcast(irc_line)
        except KeyboardInterrupt:
            for name, connection in self.connections.iteritems():
                self.send(connection.name,
                        'QUIT :%s' % getattr(self, 'exit_message', ''))
                connection.close()
            sys.exit('Bye.')


if __name__ == '__main__':
    robot = Robot()
    robot.run()
