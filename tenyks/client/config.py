import os
from os.path import abspath, join, dirname
import sys
import logging.config
from logging.handlers import SysLogHandler

from jinja2 import Template

from tenyks.module_loader import make_module_from_file


CLIENT_ROOT = abspath(dirname(__file__))


class NotConfigured(Exception):
    pass


# taken from legit (https://github.com/kennethreitz/legit)
class Settings(object):
    _singleton = {}

    # attributes with defaults
    __attrs__ = tuple()

    def __init__(self, **kwargs):
        super(Settings, self).__init__()

        self.__dict__ = self._singleton

    def __call__(self, *args, **kwargs):
        # new instance of class to call
        r = self.__class__()

        # cache previous settings for __exit__
        r.__cache = self.__dict__.copy()
        map(self.__cache.setdefault, self.__attrs__)

        # set new settings
        self.__dict__.update(*args, **kwargs)

        return r

    def __enter__(self):
        pass

    def __exit__(self, *args):
        # restore cached copy
        self.__dict__.update(self.__cache.copy())
        del self.__cache

    def __getattribute__(self, key):
        if key in object.__getattribute__(self, '__attrs__'):
            try:
                return object.__getattribute__(self, key)
            except AttributeError:
                return None
        return object.__getattribute__(self, key)


settings = Settings()

def collect_settings(settings_path=None, client_name=None):
    errors = []
    intrl_settings = None
    if not len(sys.argv) > 1:
        message = """
You need to provide a settings module.

Use `tcmkconfig clientname > /path/to/settings.py`
        """.format(pr=CLIENT_ROOT)
        raise NotConfigured(message)
    intrl_settings = make_module_from_file('settings', sys.argv[1])

    if client_name and not getattr(settings, 'CLIENT_NAME', None):
        setattr(settings, 'CLIENT_NAME', client_name)

    for sett in filter(lambda x: not x.startswith('__'), dir(intrl_settings)):
        setattr(settings, sett, getattr(intrl_settings, sett))

    if not hasattr(intrl_settings, 'WORKING_DIR'):
        WORKING_DIR = getattr(settings, 'WORKING_DIRECTORY_PATH',
                join(os.environ['HOME'], '.config', settings.CLIENT_NAME))
        DATA_WORKING_DIR = join(WORKING_DIR, 'data')
        try:
            os.makedirs(DATA_WORKING_DIR)
        except:
            errors.append('Could not create {0}'.format(DATA_WORKING_DIR))

        setattr(settings, 'WORKING_DIR', WORKING_DIR)
        setattr(settings, 'DATA_WORKING_DIR', DATA_WORKING_DIR)

    if not hasattr(intrl_settings, 'LOG_DIR'):
        LOG_DIR = WORKING_DIR
        setattr(settings, 'LOG_DIR', LOG_DIR)

    if not hasattr(settings, 'LOG_LEVEL'):
        if settings.DEBUG:
            settings.LOG_LEVEL = 'DEBUG'
        else:
            settings.LOG_LEVEL = 'INFO'

    if hasattr(settings, 'LOG_TO'):
        available_handlers = ('console', 'file', 'syslog')
        if settings.LOG_TO not in available_handlers:
            raise ConfigurationError('LOG_TO must be one of {handlers}'.format(
                handlers=available_handlers))
    else:
        setattr(settings, 'LOG_TO', 'console')

    SYSLOG_PATH = None
    if os.uname()[0] == 'Linux':
        SYSLOG_PATH = '/dev/log'
    elif os.uname()[0] == 'Darwin':
        SYSLOG_PATH = '/var/run/syslog'
    setattr(settings, 'SYSLOG_PATH', SYSLOG_PATH)

    if hasattr(intrl_settings, 'LOGGING_CONFIG'):
        LOGGING_CONFIG = intrl_settings.LOGGING_CONFIG
    else:
        LOGGING_CONFIG = {
            'version': 1,
            'disable_existing_loggers': True,
            'formatters': {
                'color': {
                    'class': 'tenyks.logs.ColorFormatter',
                    'format': '%(asctime)s %(name)s:%(levelname)s %(message)s'
                },
                'default': {
                    'format': '%(asctime)s %(name)s:%(levelname)s %(message)s'
                }
            },
            'handlers': {},
            'loggers': {
                settings.CLIENT_NAME: {
                    'handlers': [settings.LOG_TO],
                    'level': settings.LOG_LEVEL,
                    'propagate': True
                },
            }
        }
        if settings.LOG_TO == 'console':
            LOGGING_CONFIG['handlers']['console'] = {
                'level': settings.LOG_LEVEL,
                'class': 'logging.StreamHandler',
                'formatter': 'color'
            }
        elif settings.LOG_TO == 'syslog':
            LOGGING_CONFIG['handlers']['syslog'] = {
                'address': settings.SYSLOG_PATH,
                'level': settings.LOG_LEVEL,
                'class': 'logging.handlers.SysLogHandler',
                'formatter': 'default',
                'facility': SysLogHandler.LOG_SYSLOG,
            }
        elif settings.LOG_TO == 'file':
            LOGGING_CONFIG['handlers']['file'] = {
                'level': settings.LOG_LEVEL,
                'class': 'logging.FileHandler',
                'formatter': 'default',
                'filename': join(settings.LOG_DIR, 'tenyks.log')
            }


    setattr(settings, 'LOGGING_CONFIG', LOGGING_CONFIG)

    logging.config.dictConfig(LOGGING_CONFIG)

    return errors


def make_config():
    usage = 'tcmkconfig clientname'
    if len(sys.argv) < 2:
        print(usage)
        exit(1)
    with open(join(CLIENT_ROOT, 'settings.py.dist'), 'r') as f:
        settings_template = f.read()
    template = Template(settings_template)
    print(template.render({'client_name': sys.argv[1]}))
