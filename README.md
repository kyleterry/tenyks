<pre>
  _                   _         
 | |                 | |        
 | |_ ___ _ __  _   _| | _____  
 | __/ _ \ '_ \| | | | |/ / __| 
 | ||  __/ | | | |_| |   <\__ \ 
  \__\___|_| |_|\__, |_|\_\___/ 
                 __/ |          
                |___/           
</pre>

Tenyks is a (soon to be) really smart, service oriented, IRC bot. Tenyks
itself is kind of dumb, actually. He just relays messages from IRC to
services (which I call clients) listening on a redis pub/sub channel. Then he
listens for messages on another redis pub/sub channel coming from the services.
He then relays those messages to IRC based on the context sent from the client.

These current instructions are a WIP and I'm still heavily developing the bot.
Things will probably change.

# Table of Contents

* [Installing and Configuring](#installing-and-configuring)
    * [For hacking on Tenyks](#for-hacking-on-tenyks)
    * [For using Tenyks](#for-using-tenyks)
* [Running Tenyks](#running-tenyks)
* [Settings](#settings)
    * [SSL](#ssl)
* [Defaults sent to services](#defaults-sent-to-services)
* [Defaults needed sending to Tenyks](#defaults-needed-sending-to-tenyks)
* [FAQ](#faq)


## Installing and Configuring

Fist things first; you will need a [Redis](http://redis.io) server running
somewhere. Tenyks will need to connect to Redis. Look for instructions online
on how to install and run Redis.

### For hacking on Tenyks:

`mkvirtualenv tenyks`

`git clone https://github.com/kyleterry/tenyks.git`

`cd tenyks`

`python setup.py develop`

`tenyksmkconfig > /path/to/where/you/want/settings.py`

After running `tenyksmkconfig`, the settings in settings.py should make sense.
I have some comments in there explaining what each setting means. If you are
extending tenyks, anything added to settings.py will be loaded into the
`tenyks.config.settings` singleton and you can make things available for your
Tenyks extension.

`vim /path/to/where/you/want/settings.py # edit everything that makes sense to edit`

### For using Tenyks:

`pip install tenyks`

`tenyksmkconfig > /path/to/settings.py`

After running `tenyksmkconfig`, the settings in settings.py should make sense.
I have some comments in there explaining what each setting means. If you are
extending tenyks, anything added to settings.py will be loaded into the
`tenyks.config.settings` singleton and you can make things available for your
Tenyks extension.

`vim /path/to/settings.py`

## Running Tenyks

Running Tenyks is simple:

`tenyks /path/to/settings.py`

Not passing `tenyks` a settings module will result in Tenyks looking in
`~/.config/tenyks/settings.py` first and then `/etc/tenyks/settings.py`. If No
settings module is found, it will raise an error.

## Settings

The most important settings is, obviously, the IRC `CONNECTIONS` definition.
This tells Tenyks what IRC networks and channels it will be joining.

```python
CONNECTIONS = {
    'freenode': {
        'host': 'irc.freenode.net', # required
        'port': 6667, # required
        'retries': 5, # optional
        'password': None, # optional
        'nicks': ['tenyks', 'tenyks_'], # assert len(CONNECTION.get('nicks')) >= 1
        'ident': 'tenyks', # currently required
        'realname': 'tenyks IRC bot', # currently required
        'commands': ['/msg nickserv identify foo bar'], # optional
        'admins': ['yournick',], # optional
        'channels': ['#tenyks',], # assert len(CONNECTION.get('nicks')) >= 1
        # if your channel has a password: '#thechannel, thepassword'
        'ssl': False, # optional
        'ssl_version': 2 # optional. You should probably just remove this.
    },
}
```

`host` The address of the IRC network you want to connect to  
`port` The port of the IRC network you want to connect to.
Usually it's 6667, 6668 or 6669 for non-SSL. SSL is usually 6697 or 7000.  
`retries` Max retries Tenyks should make before giving up connecting to the
network  
`password` Use this if your network requires a password to connect. This is
not the channel password.  
`ident` http://wiki.swiftirc.net/index.php?title=Idents  
`realname` Not the nick. But probably something similar.  
`commands` If you registered the bot's nick, you might want to use this to
identify with the IRC network on connect. Ignore if you are not sure.  
`admins` A list of nicks that can control the bot admin style. It's commonly
used for services that require an admin to issue a command.  
`channels` A list of channels you want the bot to connect to. Don't forget the
prefix!  
`ssl` Set this to True if you are connecting to the network over SSL. Make sure
you use the network's SSL port.  
`ssl_version` Ignoring this is fine unless the network enforces a version of
SSL.  

Next most important is `REDIS_CONNECTION`. This is the server Tenyks will be
using to communicate with services. All messages Tenyks sees are sent to
Redis.

```python
REDIS_CONNECTION = {
    'host': 'localhost',
    'port': 6379,
    'db': 0,
    'password': None,
}
```

`MIDDLEWARE` is a WIP. Please ignore this setting for now.

You can set the `WORKING_DIR` and `DATA_WORKING_DIR` (these settings might be
deprecated soon).

`BROADCAST_TO_SERVICES_CHANNEL` is the Redis pubsub channel that services listen
on for messages from IRC that Tenyks relays.

`BROADCAST_TO_ROBOT_CHANNEL` is the Redis pubsub channel that Tenyks core listens
on for messages from services going to IRC. Tenyks will relay those too.

`LOGGING_DIR` is the directory where the tenyks log file will go. I suggest
logrotated.

### SSL

Tenyks supports connecting over SSL. See example settings. Currently there is
not support for self-signed certificates. This is coming.

## Defaults sent to services

```python
{
    'target': u'#test',
    'mask': '~kyle@localhost.localdomain',
    'direct': True,
    'nick': u'kyle',
    'host': u'localhost.localdomain',
    'full_message': u'tenyks: hows it going?',
    'user': u'~kyle',
    'from_channel': True,
    'payload': u'hows it going?'
}
```

`target` - Either a channel or the nick of the bot.  
`mask` - A users host as they are connected to the server.  
`direct` - Whether or not the message was directly to the bot or not.  
`nick` - The nick that the message was sent from.  
`host` - The users host.  
`full_message` - Full, unparsed message. (e.g. tenyks: Hello, world!)  
`user` - The username for the person sending the message.  
`from_channel` - Whether it was sent to a channel or as a private message to the bot.  
`payload` - The actual payload you should care about. It's the message the user
intended us to see. (e.g. Hello, world!)

## Defaults needed sending to Tenyks

```python
{
    'payload': 'this is a message to IRC',
    'target': '#test',
    'command': 'PRIVMSG',
    'connection': 'freenode',
}
```

`payload` - The message you want tenyks to send to IRC.  
`target` - The channel or user you want tenyks to send it to. IRC calls this
the target.  
`connection` - The connection where the target is.  
`command` - This is almost always PRIVMSG. PRIVMSG is actually the only thing
tenyks handles right now.

When making a Tenyks client, it's easiest to just use the tenyksclient package
in Pypi. It will handle most of the work for you and you can just think about
you client logic instead of getting your program to interface right with Tenyks.

With that said, you can write your clients in any programming language you
want. As long as you can publish a message to a Redis channel, and subscribe
to a channel on Redis, you are golden.

## FAQ

Q: Why did you make Tenyks?  
A: For shits, giggles and lols. And because I wanted to play with Gevent.

Q: How did you come up with the idea?  
A: I didn't per se. My old co-worker had an idea for a service oriented IRC bot
that used Redis pub/sub. He built that and I hacked around on it. I then
decided I wanted to play with Gevent and that led to the creation of the
evented core of Tenyks. I have a lot of experience building IRC bots because
it's one of the things I do for fun when I'm learning a new programming
language. Tenyks just became a solid piece of code and my friends liked it.

Q: Tenyks is a lot like github's HUBOT!  
A: Yep.

Q: Why can't I just use HUBOT?  
A: You can. You can also use Tenyks. Or two Tenyks'. Or two HUBOTs... Or even
two HUBOTs, an eggdrop and 4 Tenyks'. Does it matter?
