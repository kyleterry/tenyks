import os
from os.path import join

import settings


class ConfigError(Exception):
    pass 


if not getattr(settings, 'CONNECTIONS', None):
    raise ConfigError('CONNECTIONS must be set in settings.py')
CONNECTIONS = settings.CONNECTIONS

if not getattr(settings, 'REDIS_CONNECTION', None):
    raise ConfigError('REDIS_CONNECTION must be set in settings.py')
REDIS_CONNECTION = settings.REDIS_CONNECTION

WORKING_DIR = getattr(settings, 'WORKING_DIRECTORY_PATH',
        join(os.environ['HOME'], '.tenyks'))
DATA_WORKING_DIR = join(WORKING_DIR, 'data')


BROADCAST_TO_SERVICES_CHANNEL = getattr(settings,
    'BROADCAST_TO_SERVICES_CHANNEL', 'tenyks.services.broadcast_to')

BROADCAST_TO_ROBOT_CHANNEL = getattr(settings,
    'BROADCAST_TO_ROBOT_CHANNEL', 'tenyks.robot.broadcast_to')
