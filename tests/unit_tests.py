import time
from unittest import TestCase

import gevent
import gevent.server
from nose import SkipTest

from tenyks.core import Robot, Connection


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
