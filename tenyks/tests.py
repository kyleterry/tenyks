from unittest2 import TestCase


class TenyksTests(TestCase):

    def test_connection_config_validation(self):
        good_config = {
            'host': 'localhost',
            'port': 6667,
            'password': None,
            'nick': 'tenyks',
            'ident': 'tenyks',
            'realname': 'tenyks IRC bot',
            'admins': ['testnick',],
            'ssl': False,
            'channels': ['#test',],
        }

        bad_config = {
            'host': 'localhost',
            'port': 6667,
            'password': None,
            'nick': 'tenyks',
            'ident': 'tenyks',
            'realname': 'tenyks IRC bot',
            'admins': ['testnick',],
            'ssl': False,
            'channels': ['#test',],
        }

        assert False, 'todo'
