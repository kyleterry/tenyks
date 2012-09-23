from unittest import TestCase

from tenyks.core import Robot

class TenyksTestCase(TestCase):
    
    def test_can_make_irc_connection(self):
        tenyks = Robot()
