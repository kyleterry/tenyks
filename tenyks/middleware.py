from tenyks.utils import parse_irc_prefix
import re


def irc_parse(connection, data):
	raw = data
	command_re = r"^(:(?P<prefix>\S+) )?(?P<cmd>\S+)( (?!:)(?P<args>.+?))?( :(?P<trail>.+))?$"
	match_obj = re.match(command_re, data)
	data = {
		"prefix": match_obj.group("prefix"),
		"command": match_obj.group("cmd"),
		"args": match_obj.group("args"),
		"trail": match_obj.group("trail"),
	}
	print raw
	print data
	return data

def irc_extract(connection, data):
    if data['command'] == 'PRIVMSG':
        nick, user, host = parse_irc_prefix(data['prefix'])
        target = data['args']
        message = data['trail']
        data['nick'] = nick
        data['user'] = user
        data['host'] = host
        data['mask'] = '{user}@{host}'.format(user=user, host=host)
        data['connection'] = connection.name
        data['target'] = target
        if target.startswith('#'):
            data['from_channel'] = True
        else:
            data['from_channel'] = False
        data['full_message'] = message
        data['direct'] = (
                message.startswith(connection.config['nick']) or
                not data['from_channel'])
        if data['direct'] and data['from_channel']:
            data['payload'] = ' '.join(message.split()[1:])
        else:
            data['payload'] = data['full_message']
    return data

def irc_autoreply(connection, data):
	pass

def admin_middlware(connection, data):
    conf = connection.config
    print data, conf
    data['admin'] = data['nick'] in conf['admins']
    return data

CORE_MIDDLEWARE = (
    irc_parse,
    irc_extract,
    irc_autoreply,
    admin_middlware
)
