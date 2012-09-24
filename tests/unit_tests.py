import time
from unittest import TestCase

import gevent
import gevent.server

from tenyks.core import Robot, Connection


class TestServer(gevent.server.StreamServer):

    def handle(self, socket, address):
        socket.sendall('testing the connection\r\n')

class ConnectionTestCase(TestCase):
    
    def test_can_make_connection(self):
        server = TestServer(('127.0.0.1', 0))
        server.start()
        client_config = {'host': '127.0.0.1', 'port': server.server_port}
        client = Connection('test client', client_config)
        client.connect()
        assert client.socket_connected
        #client = gevent.socket.create_connection(('127.0.0.1', server.server_port))
        response = client.input_queue.get()
        assert response == 'testing the connection'
        server.stop()

    def test_can_reconnect(self):
        server = TestServer(('127.0.0.1', 0))
        server.start()
        client_config = {'host': '127.0.0.1', 'port': server.server_port}
        client = Connection('test client', client_config)
        client.connect()
        assert client.socket_connected
        server.stop()
        time.sleep(2)
        assert not client.socket_connected
        client.reconnect()
        assert client.socket_connected



class RobotTestCase(TestCase):
    
    def test_can_make_irc_connection(self):
        tenyks = Robot()
