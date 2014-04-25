import gevent.monkey
gevent.monkey.patch_all()

import json
import re

import gevent
import redis

import logging

from .config import settings, collect_settings


class Client(object):

    irc_message_filters = {}
    name = None
    direct_only = False

    def __init__(self, name):
        self.channels = [settings.BROADCAST_TO_CLIENTS_CHANNEL]
        self.name = name.lower().replace(' ', '')
        if self.irc_message_filters:
            self.re_irc_message_filters = {}
            for name, regexes in self.irc_message_filters.iteritems():
                if not name in self.re_irc_message_filters:
                    self.re_irc_message_filters[name] = []
                if isinstance(regexes, basestring):
                    regexes = [regexes]
                for regex in regexes:
                    if isinstance(regex, str) or isinstance(regex, unicode):
                        self.re_irc_message_filters[name].append(
                            re.compile(regex).match)
                    else:
                        self.re_irc_message_filters[name].append(regex)
        if hasattr(self, 'recurring'):
            gevent.spawn(self.run_recurring)
        self.logger = logging.getLogger(self.name)

    def run_recurring(self):
        self.recurring()
        recurring_delay = getattr(self, 'recurring_delay', 30)
        gevent.spawn_later(recurring_delay, self.run_recurring)

    def run(self):
        r = redis.Redis(**settings.REDIS_CONNECTION)
        pubsub = r.pubsub()
        pubsub.subscribe(self.channels)
        for raw_redis_message in pubsub.listen():
            try:
                if raw_redis_message['data'] != 1L:
                    data = json.loads(raw_redis_message['data'])
                    if self.direct_only and not data.get('direct', None):
                        continue
                    if self.irc_message_filters and 'payload' in data:
                        name, match = self.search_for_match(data['payload'])
                        ignore = (hasattr(self, 'pass_on_non_match')
                                  and self.pass_on_non_match)
                        if match or ignore:
                            self.delegate_to_handle_method(data, match, name)
                    else:
                        gevent.spawn(self.handle, data, None, None)
            except ValueError:
                self.logger.info('Invalid JSON. Ignoring message.')

    def search_for_match(self, message):
        for name, regexes in self.re_irc_message_filters.iteritems():
            for regex in regexes:
                match = regex(message)
                if match:
                    return name, match
        return None, None

    def delegate_to_handle_method(self, data, match, name):
        if hasattr(self, 'handle_{name}'.format(name=name)):
            callee = getattr(self, 'handle_{name}'.format(name=name))
            gevent.spawn(callee, data, match)
        else:
            gevent.spawn(self.handle, data, match, name)

    def handle(self, data, match, filter_name):
        raise NotImplementedError('`handle` needs to be implemented on all '
                                  'Client subclasses.')

    def send(self, message, data=None):
        r = redis.Redis(**settings.REDIS_CONNECTION)
        broadcast_channel = settings.BROADCAST_TO_ROBOT_CHANNEL
        if data:
            to_publish = json.dumps({
                'command': data['command'],
                'client': self.name,
                'payload': message,
                'target': data['target'],
                'connection': data['connection']
            })
        r.publish(broadcast_channel, to_publish)


def run_client(client_class):
    errors = collect_settings()
    client_instance = client_class(settings.CLIENT_NAME)
    for error in errors:
        client_instance.logger.error(error)
    try:
        client_instance.run()
    except KeyboardInterrupt:
        logger = logging.getLogger(client_instance.name)
        logger.info('exiting')
