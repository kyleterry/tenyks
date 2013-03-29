Tenyks
======

Tenyks is a (soon to be) really smart, service oriented, IRC bot. Tenyks
itself is kind of dumb, actually. He just relays messages from IRC to
services (which I call clients) listening on a redis pub/sub channel. Then he
listens for messages on another redis pub/sub channel coming from the services.
He then relays those messages to IRC based on the context sent from the client.

These current instructions are a WIP and I'm still heavily developing the bot.
Things will probably change.

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

`vim /path/to/where/you/want/settings.py # edit everything that makes sense to edit`

### For using Tenyks:

`pip install tenyks==0.1.20`

`tenyksmkconfig > /path/to/settings.py`

`vim /path/to/settings.py`

## Running Tenyks

Running Tenyks is simple:

`tenyks /path/to/settings.py`

## Default parameters sent to Redis from an IRC message:

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
`payload` - The actual payload you should care about. It's the message the user intended us to see. (e.g. Hello, world!)  

## Parameters needed when sending tenyks a message via Redis:

```python
{
    'payload': 'this is a message to IRC',
    'target': '#test',
    'command': 'PRIVMSG',
    'connection': 'freenode',
}
```

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
