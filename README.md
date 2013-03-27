Tenyks
======

Tenyks is a (soon to be) really smart, service oriented, IRC bot.

# Install Him

These are just development installation instructions. Production will come
soon when I am happy with it.

`$ mkvirtualenvwrapper tenyks`

`$ git clone https://github.com/kyleterry/tenyks.git`

`$ python setup.py develop`

`$ cp tenyks/settings.py{,dist}`

`$ vim tenyks/settings.py # edit everything that makes sense to edit`

`$ tenyks`

# Default parameters sent to Redis from an IRC message:

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

`target` - Either a channel or the nick of the bot.
`mask` - A users host as they are connected to the server.
`direct` - Whether or not the message was directly to the bot or not.
`nick` - The nick that the message was sent from.
`host` - The users host.
`full_message` - Full, unparsed message. (e.g. tenyks: Hello, world!)
`user` - The username for the person sending the message.
`from_channel` - Whether it was sent to a channel or as a private message to the bot.
`payload` - The actual payload you should care about. It's the message the user inteded us to see. (e.g. Hello, world!)
