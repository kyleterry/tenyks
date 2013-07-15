
class Command(object):
    template = ''

    def __init__(self, *args, **kwargs):
        self.args = args
        self.kwargs = kwargs

    def __str__(self):
        return self.template.format(*self.args, **self.kwargs)

class JOIN(Command):
    template = 'JOIN {channel} {password}'

class NICK(Command):
    template = 'NICK {nick}'

class PASS(Command):
    template = 'PASS {password}'

class PRIVMSG(Command):
    template = 'PRIVMSG {target} :{message}'

class QUIT(Command):
    template = 'QUIT :{message}'

class USER(Command):
    template = 'USER {ident} {host} bla :{realname}'


