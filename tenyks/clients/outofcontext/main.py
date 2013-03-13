import random

import gevent.monkey
gevent.monkey.patch_all()

from tenyks.client import Client, run_client


class TenyksOutOfContext(Client):

    def __init__(self):
        super(TenyksOutOfContext, self).__init__()
        self.messages = {}
        self.choices = [
            'I burnt my tongue.',
            'So that\'s where babies come from...',
            'I\'m sleepy',
            'If a farting wiener dog ever falls into a deep puddle we are all fucked.',
        ]

    def handle(self, data, match, filter_name):
        if data['irc_channel'].startswith('#'):
            try:
                count = self.messages[data['connection_name']][data['irc_channel']]
                self.messages[data['connection_name']][data['irc_channel']] = count + 1
            except:
                count = 0
                self.messages[data['connection_name']] = {
                    data['irc_channel']: count}
            self.logger.debug('{conn}:{channel} message count: {count}'.format(
                conn=data['connection_name'], channel=data['irc_channel'],
                count=self.messages[data['connection_name']][data['irc_channel']]))
            self.last_nick = data['nick_from']
            if count + 1 >= 10:
                self.send(random.choice(self.choices), data)
                self.messages[data['connection_name']][data['irc_channel']] = 0


if __name__ == '__main__':
    out_of_context = TenyksOutOfContext()
    run_client(out_of_context)
