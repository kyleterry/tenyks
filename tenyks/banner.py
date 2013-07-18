import pkg_resources

from clint.textui import colored
import redis

from tenyks.config import settings

version = pkg_resources.require('tenyks')[0].version

BANNER = [
    "  _                   _         ",
    " | |                 | |        ",
    " | |_ ___ _ __  _   _| | _____  ",
    " | __/ _ \ '_ \| | | | |/ / __| ",
    " | ||  __/ | | | |_| |   <\__ \ ",
    "  \__\___|_| |_|\__, |_|\_\___/ ",
    "                 __/ |          ",
    "                |___/           ",
]

version_string = "Version: {version}\n\n"


NETWORKS = """\
[Configuration]
\tDebug: {debug}
\tRedis: redis://{redis_host}:{redis_port} [{redis_status}]

[Networks]
{networks}
"""


def get_redis_status():
    rds = redis.Redis(**settings.REDIS_CONNECTION)
    try:
        rds.ping()
        status = colored.green('OK')
    except redis.ConnectionError:
        status = colored.red('FAILED')
    return status


def startup_banner():
    banner = '\n'.join(BANNER) + '\n\n'
    banner = banner + version_string + NETWORKS
    networks = []
    for name, connection in settings.CONNECTIONS.items():
        if 'ssl' in connection and connection['ssl']:
            name = name + ' (SSL)'
        networks.append('\t' + name)
    banner = banner.format(
        version=colored.green(version), networks='\n'.join(networks),
        debug=colored.green(
            str((settings.DEBUG if hasattr(settings, 'DEBUG') else False))),
        redis_host=settings.REDIS_CONNECTION['host'],
        redis_port=settings.REDIS_CONNECTION['port'],
        redis_status=get_redis_status())
    return banner
