import os
import sys
import logging
import gevent
from gevent import monkey
from gevent import socket, queue
from gevent.ssl import wrap_socket
from ssl import CERT_NONE
import time
from threading import Thread, Lock


class Connection(object):

    def __init__(self):
        self.greelets = []
        self.socket = self.create_socket()

    def create_socket(self):
        """ This method exists to create ssl sockets, also
        """
        return socket.Socket()

    def connect(self):
        pass

    def recv_loop(self):
        pass

    def send_loop(self):
        pass


class Line(object):
    pass


class Tenyks(object):
    pass
