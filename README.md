[![Build Status](https://travis-ci.org/kyleterry/tenyks.svg?branch=master)](https://travis-ci.org/kyleterry/tenyks)
```Textile
  _                   _         
 | |                 | |        
 | |_ ___ _ __  _   _| | _____  
 | __/ _ \ '_ \| | | | |/ / __| 
 | ||  __/ | | | |_| |   <\__ \  The IRC bot for hackers.
  \__\___|_| |_|\__, |_|\_\___/ 
                 __/ |          
                |___/           
```

Tenyks is a service oriented IRC bot rewritten in Go. Service/core
communication is handles by Redis Pub/Sub via json payloads.

The core acts like a relay between IRC channels and remote services. When a
message comes in from IRC, that message is turned into a json data structure,
then sent over the pipe on a Pub/Sub channel that services can subscribe to.
Services then parse or pattern match the message, and possibly respond back via
the same method.

This design, while not anything new, is very flexible because one can write
their service in any number of languages. The current service implementation
used for proof of concept is written in Python. You can find that
[here](./legacy/tenyks/client). It's also beneficial because you can take down
or bring up services without the need to restart the bot or implement a
complicated hot pluggable core. Services that crash also don't run the risk of
taking everything else down with it.

## Installation and building

Since Tenyks is pretty new, you will need to build the bot yourself. Step 1 is
making sure you have a redis-server running. I won't go into detail there as I'm
sure you can figure it out. Step 2 is making sure you have all the Go
dependencies installed. For instance on Debian you would run `sudo apt-get
install golang`.

### Building

You can build with the make, which calls [Go tool](http://golang.org/cmd/go/)
things.

```bash
git clone https://github.com/kyleterry/tenyks
cd tenyks
make # tenyks will be in ./bin
sudo make install
```

`tenyks` should now be in `/usr/local/bin/tenyks` (or whatever you chose for
your PREFIX)

### Uninstall

Why would you ever want to do that?

```bash
cd /path/to/tenyks
sudo make uninstall
```

### Configuration

Configuration is just json. The included example contains everything you need to
get started. You just need to swap out the server information.

```bash
cp config.json.example ${HOME}/tenyks-config.json
```

## Running

`tenyks ${HOME}/tenyks-config.json`

If a config file is excluded when running, Tenyks will look for configuration
in `/etc/tenyks/config.json` first, then
`${HOME}/.config/tenyks/config.json` then it will give up. These are defined
in tenyks/tenyks.go and added with ConfigSearch.AddPath(). If you feel more
paths should be searched, please feel free to add it and submit a pull request.

## Testing

I'm a horrible person. ~~There aren't tests yet. I'll get right on this...~~.
There are only a few tests.

## Services

### To Services

Example JSON payload sent to services:

```json
{
    "target":"#tenyks",
    "command":"PRIVMSG",
    "mask":"unaffiliated/vhost-",
    "direct":true,
    "nick":"vhost-",
    "host":"unaffiliated/vhost-",
    "fullmsg":":vhost-!~vhost@unaffiliated/vhost- PRIVMSG #tenyks :tenyks-demo: weather 97217",
    "full_message":":vhost-!~vhost@unaffiliated/vhost- PRIVMSG #tenyks :tenyks-demo: weather 97217",
    "user":"~vhost",
    "fromchannel":true,
    "from_channel":true,
    "connection":"freenode",
    "payload":"weather 97217",
    "meta":{
        "name":"Tenyks",
        "version":"1.0"
    }
}
```

fullmsg, full_message and fromchannel from_channel are for backwards
compatibility with older services.

### To Tenyks for IRC

Example JSON response from a service to Tenyks destined for IRC

```json
{
    "target":"#tenyks",
    "command":"PRIVMSG",
    "fromchannel":true,
    "from_channel":true,
    "connection":"freenode",
    "payload":"Portland, OR is 63.4 F (17.4 C) and Overcast; windchill is NA; winds are Calm",
    "meta":{
        "name":"TenyksWunderground",
        "version":"1.1"
    }
}
```

### Service Registration

```json
{
    "command":"REGISTER",
    "meta":{
        "name":"TenyksWunderground",
        "version":"1.1"
    }
}
```

### Service going offline

```json
{
    "command":"BYE",
    "meta":{
        "name":"TenyksWunderground",
        "version":"1.1"
    }
}
```

### Lets make a service!

This service is in python and uses the
[tenyks-service](https://github.com/kyleterry/tenyks-service) package. You can
install that with pip: `pip install tenyksservice`.

```python
from tenyksservice import TenyksService, run_service


class Hello(TenyksService):
    direct_only = True
    irc_message_filters = {
        'hello': [r"^(?i)(hi|hello|sup|hey), I'm (?P<name>(.*))$"]
    }

    def handle_hello(self, data, match):
        name = match.groupdict()['name']
        self.logger.debug('Saying hello to {name}'.format(name=name))
        self.send('How are you {name}?!'.format(name=name), data)


def main():
    run_service(Hello)


if __name__ == '__main__':
    main()
```

Okay, we need to generate some settings for our new service.

```bash
tenyks-service-mkconfig hello >> hello_settings.py
```

Now lets run it: `python main.py hello_settings.py`

If you now join the channel that tenyks is in and say "tenyks: hello, I'm Alice"
then tenyks should respond with "How are you Alice?!".

## Help me

I'm a new Go programmer. Surely there's some shitty shit in here. You can help
me out by creating an issue on Github explaining how dumb I am. Or you can patch
the dumbness, make a pull request and tell me what I did wrong and why you made
the change you did. I'm open to criticism as long as it's done in a respectful
and "I'm teaching you something new" kind of way.

## The future

I'm using Redis pub/sub right now. It's pretty coupled to that, but I plan to
implement a more pluggable communication backend.

I also plan on trying to implement pub/sub over web sockets that will allow
for communication over TLS and have authentication baked in. One use-case is
giving your friends authentication tokens they can use to authenticate their
services with your IRC bot. If someone is trouble, you cut them off.

I'm also planning a 2.0 branch which will use ZeroMQ instead of Redis. I like
embedded stuff.

## Credit where credit is due

Service oriented anything isn't new. This idea came from an old
[coworker](https://github.com/wraithan) of mine. I just wanted to do something
slightly different. There are also plenty of other plugin style bots (such as
hubot and eggdrop). Every open source project needs love, so check those out as
well.
