import json
import re

import gevent
from gevent import queue
import redis

import logging
logger = logging.getLogger()

import tenyks.config as config


CLIENT_SERVICE_STATUS_OFFLINE = 0
CLIENT_SERVICE_STATUS_ONLINE = 1


class Client(object):

    message_filter = None
    name = None
    direct_only = False

    def __init__(self):
        self.input_queue = queue.Queue()
        self.output_queue = queue.Queue()
        if not self.name:
            self.name = self.__class__.__name__.lower()
        else:
            self.name = self.name.lower()
        if self.message_filter:
            self.re_message_filter = re.compile(self.message_filter).match

    def run(self):
        r = redis.Redis(**config.REDIS_CONNECTION)
        pubsub = r.pubsub()
        pubsub.subscribe(config.BROADCAST_TO_SERVICES_CHANNEL)
        for raw_redis_message in pubsub.listen():
            logger.debug(json.dumps(raw_redis_message))
            try:
                if raw_redis_message['data'] != 1L:
                    data = json.loads(raw_redis_message['data'])
                    if self.direct_only and not data['direct']:
                        continue
                    if self.message_filter:
                        match = self.re_message_filter(data['payload'])
                        if match:
                            gevent.spawn(self.handle, data, match)
                    else:
                        gevent.spawn(self.handle, data, None)
            except ValueError:
                logger.info(
                        '{name}.run: invalid JSON. Ignoring message.'.format(
                                    name=self.__class__.__name__))

    def handle(self, data, match):
        raise NotImplementedError('`handle` needs to be implemented on all '
                                  'Client subclasses.')

    def send(self, message, data=None):
        r = redis.Redis(**config.REDIS_CONNECTION)
        broadcast_channel = config.BROADCAST_TO_ROBOT_CHANNEL
        if data:
            to_publish = json.dumps({
                'version': 1,
                'type': 'privmsg',
                'client': self.name,
                'payload': message,
                'irc_channel': data['irc_channel'],
                'connection_name': data['connection_name']
            })
        r.publish(broadcast_channel, to_publish)


def run_client(service_instance):
    try:
        service_instance.run()
    except KeyboardInterrupt:
        logger.info('%s client: exiting' % service_instance.name)
    finally:
        pass
        #with open(service_instance.log_file, 'a+') as f:
        #    f.write('Shutting down')
        #service_instance.send_status_update(CLIENT_SERVICE_STATUS_OFFLINE)
