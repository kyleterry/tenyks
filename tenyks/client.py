import re
import json
import redis
from os.path import join

import gevent
from gevent import queue

import settings
from constants import PING, PONG

CLIENT_TYPE_ONE_TIME = 1
CLIENT_TYPE_SERVICE = 2
CLIENT_SERVICE_STATUS_ONLINE = 1
CLIENT_SERVICE_STATUS_OFFLINE = 2


class TenyksClientInitError(Exception):
    pass


class TenyksClient(object):

    def __init__(self):
        if not getattr(self, 'client_type'):
            raise TenyksClientInitError(
                    'You must subclass with a client_type attribute')
        if not getattr(self, 'service_name'):
            raise TenyksClientInitError(
                    'You must subclass with a service_name attribute')
        self.r = redis.Redis(**settings.REDIS_CONNECTION)
        if self.client_type == CLIENT_TYPE_SERVICE:
            log_directory = getattr(settings, 'LOG_DIRECTORY', '/tmp/tenyks')
            self.log_file = join(log_directory, 'tenyks-service-%s.log' % self.service_name)
            self.input_queue = queue.Queue()
            with open(self.log_file, 'a+') as f:
                f.write('Starting up')
            gevent.spawn(self.pub_sub_loop)
            self.send_status_update(CLIENT_SERVICE_STATUS_ONLINE)

    def pub_sub_loop(self):
        pubsub = self.r.pubsub()
        broadcast_channel = getattr(settings, 'BROADCAST_TO_SERVICES_CHANNEL',
            'tenyks.services.broadcast_to')
        pubsub.subscribe(broadcast_channel)
        for raw_redis_message in pubsub.listen():
            if raw_redis_message['data'] != 1L: 
                try:
                    data = json.loads(raw_redis_message['data'])
                    if data['version'] == 1:
                        if data['type'] == 'privmsg':
                            self.input_queue.put(data)
                        elif (self.client_type == CLIENT_TYPE_SERVICE and 
                            data['type'] == PING and
                            data['data']['name'] == self.service_name):
                            self.handle_pint(data)
                except:
                    print 'Failed to parse JSON.'

    def handle_ping(self, data):
        print 'responding to PING on %s' % (
                data['data']['channel'])
        ping_response_key = getattr(settings, 'SERVICES_PING_RESPONSE_KEY',
                'tenyks.services.%s.ping_response' % self.service_name)
        self.r.rpush(ping_response_key, PONG)

    def publish(self, to_publish, broadcast_channel=None):
        if not broadcast_channel:
            broadcast_channel = getattr(settings, 'BROADCAST_TO_ROBOT_CHANNEL',
                'tenyks.robot.broadcast_to')
        self.r.publish(broadcast_channel, to_publish)

    def send_status_update(self, status):
        to_publish = json.dumps({
            'version': 1,
            'type': 'client_status',
            'data': {
                'status': status,
                'name': self.service_name
            }
        })
        self.publish(to_publish)

    def hear(self, expression):
        """
        Generator that compiles a regular expression and returns a re result
        """
        listening = re.compile(expression).match
        while True:
            result = listening(self.input_queue.get())
            if result:
                yield result

    def run(self):
        raise NotImplementedError('You must subclass with a `run` method.')


def run_service(service_instance):
    try:
        service_instance.run()
    finally:
        with open(service_instance.log_file, 'a+') as f:
            f.write('Shutting down')
        service_instance.send_status_update(CLIENT_SERVICE_STATUS_OFFLINE)
