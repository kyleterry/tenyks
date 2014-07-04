import logging

from termcolors import colorize


COLORS = {
    'DEBUG': ('white', ''),
    'INFO': ('green', ''),
    'WARNING': ('yellow', ''),
    'CRITICAL': ('red', ''),
    'ERROR': ('white', 'red'),
}


class ColorFormatter(logging.Formatter):

    def format(self, record):
        if record.levelname in COLORS:
            colors = COLORS[record.levelname]
            kwargs = {'fg': colors[0]}
            if colors[1]:
                kwargs['bg'] = colors[1]
            record.msg = colorize(record.msg, **kwargs)
        if record.levelname == 'INFO':
            if 'success' in record.msg or 'OK' in record.msg:
                record.msg = colorize(record.msg, fg='green')
        return super(ColorFormatter, self).format(record)
