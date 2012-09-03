import logging
from logging import DEBUG, INFO, WARNING, ERROR


BLACK, RED, GREEN, YELLOW, BLUE, MAGENTA, CYAN, WHITE = range(8)

COLORS = {
    'WARNING': YELLOW,
    'INFO': WHITE,
    'DEBUG': BLUE,
    'CRITICAL': YELLOW,
    'ERROR': RED
}

#The background is set with 40 plus the number of the color, and the foreground with 30

#These are the sequences need to get colored ouput
RESET_SEQ = '\033[0m'
COLOR_SEQ = '\033[1;%dm'
BOLD_SEQ = '\033[1m'

COLOR_GREEN = COLOR_SEQ % GREEN
OK = '[%sOK%s]' % (COLOR_GREEN, RESET_SEQ)
COLOR_RED = COLOR_SEQ % RED
FAILED = '[%sFAILED%s]' % (COLOR_RED, RESET_SEQ)


#def formatter_message(message, use_color=True):
#    if use_color:
#        message = message.replace('$RESET', RESET_SEQ).replace('$BOLD', BOLD_SEQ)
#    else:
#        message = message.replace('$RESET', '').replace('$BOLD', '')
#    return message
#
#
#class ColoredFormatter(logging.Formatter):
#    def __init__(self, message, use_color=True):
#        logging.Formatter.__init__(self, message)
#        self.use_color = use_color
#
#    def format(self, record):
#        levelname = record.levelname
#        if self.use_color and levelname in COLORS:
#            levelname_color = COLOR_SEQ % (30 + COLORS[levelname]) + levelname + RESET_SEQ
#            record.levelname = levelname_color
#        return logging.Formatter.format(self, record)
#
#
#class ColoredLogger(logging.Logger):
#    FORMAT = '[$BOLD%(name)-20s$RESET][%(levelname)-18s]  %(message)s ($BOLD%(filename)s$RESET:%(lineno)d)'
#    COLOR_FORMAT = formatter_message(FORMAT, True)
#    def __init__(self, name):
#        logging.Logger.__init__(self, name, logging.DEBUG)                
#
#        color_formatter = ColoredFormatter(self.COLOR_FORMAT)
#
#        console = logging.StreamHandler()
#        console.setFormatter(color_formatter)
#
#        self.addHandler(console)
#        return


logging.basicConfig(format='%(message)s')
logger = logging.getLogger()

logging.basicConfig(format='[%(levelname)s] %(message)s')
debug_logger = logging.getLogger()


def make_log(message, type=INFO):
    logger.info(message)


def make_debug_log(message):
    debug_logger.debug(message)
