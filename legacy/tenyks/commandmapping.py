class CommandNotFound(KeyError):
    pass


class Command(object):

    template = None

    def __init__(self, command_string):
        if command_string.startswith('/'):
            command_string = command_string.lstrip('/')
        self.command_string = command_string
        self.parts = command_string.split()

    def to_irc(self):
        raise NotImplemented('You need to define a `to_irc` method')


class PrivmsgCommand(Command):

    template = 'PRIVMSG {target} :{message}'

    def to_irc(self):
        message = ' '.join(self.parts[2:])
        return self.template.format(
            target=self.parts[1], message=message)


class NickCommand(Command):

    template = 'NICK {nick}'

    def to_irc(self):
        return self.template.format(nick=self.parts[1])


COMMAND_MAP= {
    'msg': PrivmsgCommand,
    'nick': NickCommand,
}


def command_parser(command_str):
    if command_str.startswith('/'):
        command_parts = command_str.lstrip('/').split()
        try:
            command_cls = COMMAND_MAP[command_parts[0].lower()]
        except KeyError:
            raise CommandNotFound('/{command}... not found'.format(
                command=command_parts[0]))
        command = command_cls(command_str)
        return command.to_irc()
    else:
        return command_str
