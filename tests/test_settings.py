# Tenyks settings

# Anything added to this file will be included in the tenyks.config.settings
# singleton.

DEBUG = False

##############################################################################
# The following setting defines a dictionary of IRC connections.
#
# This setting is required.

CONNECTIONS = {
    'network1': {
        'host': 'localhost',
        'port': 6667,
        'password': None,
        'nick': 'tenyks',
        'ident': 'tenyks',
        'realname': 'tenyks IRC bot',
        'admins': ['vhost-',],
        'channels': ['#test',], # if your channel has a password: '#thechannel, thepassword'
        #'ssl': False, # not supported yet.
    },
}
##############################################################################


##############################################################################
# The following setting defines a dictionary containing Redis connection
# information used when connecting to the Redis server.
#
# This setting is required.

REDIS_CONNECTION = {
    'host': 'localhost',
    'port': 6379,
    'db': 0,
    'password': None,
}
##############################################################################


##############################################################################
# MIDDELWARE is a tuple of callables that can manipulate the data dictionary
# each message is generated before it gets published to Redis.
#
# This setting is optional

MIDDLEWARE = (
    # Custom Middlware goes in here.
)
##############################################################################


##############################################################################
# The following two settings default to /home/<currentuser>/.config/tenyks
# Uncomment them if you need to set a different working directory.
# The data working directory, if not set, will default to be a subdirectory of
# the working directory.
#
# These settings are optional

# WORKING_DIR = /path/to/working/dir
# DATA_WORKING_DIR = /path/to/working/dir/data
##############################################################################


##############################################################################
# The following channel settings are for when you need to change the default
# Redis channel key for each pipe respectively. BROADCAST_TO_SERVICES_CHANNEL
# is used for messages going from Tenyks to Redis for the clients.
# BROADCAST_TO_ROBOT_CHANNEL is used for messages going from clients to Tenyks
# for IRC.
#
# These are namespaced, so you shouldn't need to change them.
#
# These settings are optional

# BROADCAST_TO_SERVICES_CHANNEL = 'tenyks.services.broadcast_to'
# BROADCAST_TO_ROBOT_CHANNEL = 'tenyks.robot.broadcast_to'
##############################################################################
