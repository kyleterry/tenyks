import time
from unittest import TestCase
from os.path import abspath, join, dirname

import gevent
import gevent.server
from nose import SkipTest

PROJECT_ROOT = abspath(dirname(__file__))
from tenyks.core import Robot, Connection
from tenyks.utils import parse_irc_message, parse_irc_prefix, parse_args
from tenyks.config import collect_settings


collect_settings(join(PROJECT_ROOT, 'test_settings.py'))


class TestServer(gevent.server.StreamServer):

    def handle(self, socket, address):
        self.socket1 = socket
        socket.sendall('testing the connection\r\n')
        gevent.spawn(self.recv)

    def recv(self):
        data = self.socket1.recv(1024)
        self.input_data = data


class ConnectionTestCase(TestCase):

    def test_can_make_connection(self):
        server = TestServer(('127.0.0.1', 0))
        server.start()
        client_config = {'host': '127.0.0.1', 'port': server.server_port}
        client = Connection('test client', client_config)
        client.connect()
        assert client.socket_connected
        response = client.input_queue.get()
        assert response == 'testing the connection'
        server.stop()

    def test_can_reconnect(self):
        raise SkipTest
        server = TestServer(('127.0.0.1', 0))
        server.start()
        client_config = {'host': '127.0.0.1', 'port': server.server_port}
        client = Connection('test client', client_config)
        client.connect()
        assert client.socket_connected
        server.stop()
        time.sleep(6)
        assert client.server_disconnect
        assert not client.socket_connected
        server.start()
        client.reconnect()
        assert client.socket_connected

    def test_can_send_data(self):
        server = TestServer(('127.0.0.1', 0))
        server.start()
        client_config = {'host': '127.0.0.1', 'port': server.server_port}
        client = Connection('test client', client_config)
        client.connect()
        assert client.socket_connected
        client.output_queue.put('testing sending')
        time.sleep(2)
        self.assertEqual(server.input_data, 'testing sending\r\n')
        server.stop()


class RobotTestCase(TestCase):

    def test_can_make_irc_connection(self):
        tenyks = Robot()


class IRCParsingTestCase(TestCase):

    def test_can_break_raw_irc_message_up(self):
        message = ':kyle!~kyle@localhost.localdomain PRIVMSG #test :tenyks: hi'
        parsed_message = parse_irc_message(message)
        assert len(parsed_message) == 3
        assert parsed_message[0] == 'kyle!~kyle@localhost.localdomain'
        assert parsed_message[1] == 'PRIVMSG'
        assert len(parsed_message[2]) == 2
        assert parsed_message[2][0] == '#test'
        assert parsed_message[2][1] == 'tenyks: hi'

    def test_can_parse_prefix(self):
        message = ':kyle!~kyle@localhost.localdomain PRIVMSG #test :tenyks: hi'
        parsed_message = parse_irc_message(message)
        parsed_prefix = parse_irc_prefix(parsed_message[0])
        assert parsed_prefix[0] == 'kyle'
        assert parsed_prefix[1] == '~kyle'
        assert parsed_prefix[2] == 'localhost.localdomain'

    def test_can_parse_args(self):
        message = ':kyle!~kyle@localhost.localdomain PRIVMSG #test :tenyks: hi'
        parsed_message = parse_irc_message(message)
        args = parse_args(parsed_message[2])
        assert args[0] == '#test'
        assert args[1] == 'tenyks: hi'
