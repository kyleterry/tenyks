import re
import json
import redis
from os.path import join
import logging
logger = logging.getLogger(__name__)

import gevent
from gevent import queue

import config
from constants import PING, PONG

CLIENT_TYPE_ONE_TIME = 1
CLIENT_TYPE_SERVICE = 2
CLIENT_SERVICE_STATUS_ONLINE = 1
CLIENT_SERVICE_STATUS_OFFLINE = 2


class TenyksClientInitError(Exception):
    pass


class TenyksClient(object):

    def __init__(self):
        if not getattr(self, 'client_name'):
            raise TenyksClientInitError(
                    'You must subclass with a client_name attribute')
        self.r = redis.Redis(**config.REDIS_CONNECTION)
        log_directory = getattr(config, 'LOG_DIR', config.WORKING_DIR)
        self.log_file = join(log_directory, 'tenyks-service-%s.log' % self.client_name)
        self.input_queue = queue.Queue()
        with open(self.log_file, 'a+') as f:
            f.write('Starting up')
        gevent.spawn(self.pub_sub_loop)
        self.send_status_update(CLIENT_SERVICE_STATUS_ONLINE)
        if getattr(self, 'hear', None):
            self.heard_queue = queue.Queue()
            gevent.spawn(self.hear)

    def pub_sub_loop(self):
        pubsub = self.r.pubsub()
        broadcast_channel = getattr(config, 'BROADCAST_TO_SERVICES_CHANNEL',
            'tenyks.services.broadcast_to')
        pubsub.subscribe(broadcast_channel)
        for raw_redis_message in pubsub.listen():
            if raw_redis_message['data'] != 1L: 
                try:
                    data = json.loads(raw_redis_message['data'])
                    if data['version'] == 1:
                        if data['type'] == 'privmsg':
                            self.input_queue.put(data)
                        elif data['type'] == PING and \
                            data['data']['name'] == self.client_name:
                            self.handle_ping(data)
                except ValueError:
                    logger.error('Failed to parse Json')

    def handle_ping(self, data):
        logger.debug('responding to PING on %s' % (
                data['data']['channel']))
        ping_response_key = getattr(config, 'SERVICES_PING_RESPONSE_KEY',
                'tenyks.services.%s.ping_response' % self.client_name)
        self.r.rpush(ping_response_key, PONG)

    def publish(self, to_publish, broadcast_channel=None):
        if not broadcast_channel:
            broadcast_channel = getattr(config, 'BROADCAST_TO_ROBOT_CHANNEL',
                'tenyks.robot.broadcast_to')
        self.r.publish(broadcast_channel, to_publish)

    def send_status_update(self, status):
        to_publish = json.dumps({
            'version': 1,
            'type': 'client_status',
            'data': {
                'status': status,
                'name': self.client_name
            }
        })
        self.publish(to_publish)

    def hear(self):
        """
        Generator that compiles a regular expression and returns a re result
        """
        listening = re.compile(self.hear).match
        while True:
            data = self.input_queue.get()
            self.heard_queue.put(data)
            print data

    def run(self):
        raise NotImplementedError('You must subclass with a `run` method.')


def run_service(service_instance):
    try:
        service_instance.run()
    except KeyboardInterrupt:
        logger.info('%s client: exiting' % service_instance.name)
    finally:
        with open(service_instance.log_file, 'a+') as f:
            f.write('Shutting down')
        service_instance.send_status_update(CLIENT_SERVICE_STATUS_OFFLINE)
