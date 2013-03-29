Tenyks
======

Tenyks is a (soon to be) really smart, service oriented, IRC bot.

## Installation

### For hacking on Tenyks:

These are just development installation instructions. Production will come
soon when I am happy with it.

Fist things first; you will need a Redis server running somewhere. Tenyks will
need to connect to Redis. Look for instructions online on how to install and
run Redis.

`$ mkvirtualenvwrapper tenyks`

`$ git clone https://github.com/kyleterry/tenyks.git`

`cd tenyks`

`$ python setup.py develop`

`$ cp tenyks/settings.py.dist /path/to/where/you/want/settings.py`

`$ vim /path/to/where/you/want/settings.py # edit everything that makes sense to edit`

`$ tenyks /path/to/where/you/want/settings.py`

### For using Tenyks:

Coming soon.

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
