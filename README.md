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

Tenyks is a computer program designed to relay messages between connections to
IRC networks and custom built services written in any number of languages.
More detailed, Tenyks is a service oriented IRC bot rewritten in Go.
Service/core communication is handled by ZeroMQ 4 PubSub via json payloads.

The core acts like a relay between IRC channels and remote services. When a
message comes in from IRC, that message is turned into a json data structure,
then sent over the pipe on a Pub/Sub channel that services can subscribe to.
Services then parse or pattern match the message, and possibly respond back via
the same method.

This design, while not anything new, is very flexible because one can write
their service in any number of languages. The current service implementation
used for proof of concept is written in Python. You can find that
[here](https://github.com/kyleterry/tenyks-service). It's also beneficial because you can take down
or bring up services without the need to restart the bot or implement a
complicated hot pluggable core. Services that crash also don't run the risk of
taking everything else down with it.

## Installation and whatnot

### Building

Current supported Go version is 1.5. All packages are vendored with Godep and
stored in the repository. I update these occasionally. Make sure you have a
functioning Go 1.5 environment with `GO15VENDOREXPERIMENT=1`.

1. Install ZeroMQ4 (reference your OSs package install documentation) and make
   sure libzmq exists on the system.
1. `go get github.com/kyleterry/tenyks`
1. `cd ${GOPATH}/src/github.com/kyleterry/tenyks`
1. `make` - this will run tests and build
1. `sudo make install` - otherwise you can find it in `./bin/tenyks`
1. `cp config.json.example config.json`
1. Edit `config.json` to your liking.

### Uninstall

Why would you ever want to do that?

```bash
cd ${GOPATH}/src/github.com/kyleterry/tenyks
sudo make uninstall
```

### Binary Release

You can find binary builds on [bintray](https://dl.bintray.com/kyleterry/tenyks/).

I cross compile for Linux {arm,386,amd64} and Darwin {386,amd64}.

### Configuration

Configuration is just json. The included example contains everything you need to
get started. You just need to swap out the server information.

```bash
cp config.json.example ${HOME}/tenyks-config.json
```

### Running

`tenyks ${HOME}/tenyks-config.json`

If a config file is excluded when running, Tenyks will look for configuration
in `/etc/tenyks/config.json` first, then
`${HOME}/.config/tenyks/config.json` then it will give up. These are defined
in tenyks/tenyks.go and added with ConfigSearch.AddPath(). If you feel more
paths should be searched, please feel free to add it and submit a pull request.

### Vagrant

If you want to play _right fucking now_, you can just use vagrant: `vagrant up`
and then `vagrant ssh`. Tenyks should be built and available in your `$PATH`.
There is also an IRC server running you can connect to server on `192.168.33.66` with your IRC client.

Just run `tenyks & && disown` from the vagrant box and start playing.

## Testing

I'm a horrible person. ~~There aren't tests yet. I'll get right on this...~~.
There are only a few tests.

## Builtins

Tenyks comes with very few commands that the core responds to directly. You can
get a list of services and get help for those services.

`tenyks: !services` will list services that have registered with the bot
through the service [registration API.](#service-registration).  
`tenyks: !help` will show a quick help menu of all the commands available to
tenyks.  
`tenyks: !help servicename` will ask the service to sent their help message to
the user.

## Services

### Libraries
* [tenyksservice](https://github.com/kyleterry/tenyks-service) (Python)
* [quasar](https://github.com/kyleterry/quasar) (Go)

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
    "full_message":":vhost-!~vhost@unaffiliated/vhost- PRIVMSG #tenyks :tenyks-demo: weather 97217",
    "user":"~vhost",
    "from_channel":true,
    "connection":"freenode",
    "payload":"weather 97217",
    "meta":{
        "name":"Tenyks",
        "version":"1.0"
    }
}
```

### To Tenyks for IRC

Example JSON response from a service to Tenyks destined for IRC

```json
{
    "target":"#tenyks",
    "command":"PRIVMSG",
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

Registering your service with the bot will let people ask Tenyks which services
are online and available for use. Registering is not requires; anything
listening on the pubsub channel can respond without registration.

Each service should have a unique UUID set in it's REGISTER message. An example
of a valid register message is below:

```json
{
    "command":"REGISTER",
    "meta":{
        "name":"TenyksWunderground",
        "version":"1.1",
        "UUID": "uuid4 here",
        "description": "Fetched weather for someone who asks"
    }
}
```

### Service going offline

If the service is shutting down, you should send a BYE message so Tenyks doesn't
have to timeout the service after PINGs go unresponsive:

```json
{
    "command":"BYE",
    "meta":{
        "name":"TenyksWunderground",
        "version":"1.1",
        "UUID": "uuid4 here",
        "description": "Fetched weather for someone who asks"
    }
}
```

### Commands for registration that go to services

Services can register with Tenyks. This will allow you to list the services
currently online from the bot. This is not persistent. If you shut down the bot,
then all the service UUIDs that were registered go away.

The commands sent to services are:

```json
{
  "command": "HELLO",
  "payload": "!tenyks"
}
```  
`HELLO` will tell services that Tenyks has come online and they can register if
they want to.

```json
{
  "command": "PING",
  "payload": "!tenyks"
}
```  
`PING` will expect services to respond with `PONG`.

List and Help commands are coming soon.

### Lets make a service!

This service is in python and uses the
[tenyks-service](https://github.com/kyleterry/tenyks-service) package. You can
install that with pip: `pip install tenyksservice`.

```python
from tenyksservice import TenyksService, run_service, FilterChain


class Hello(TenyksService):
    irc_message_filters = {
        'hello': FilterChain([r"^(?i)(hi|hello|sup|hey), I'm (?P<name>(.*))$"],
                             direct_only=False),
        # This is will respond to /msg tenyks this is private
        'private': FilterChain([r"^this is private$"],
                             private_only=True)
    }

    def handle_hello(self, data, match):
        name = match.groupdict()['name']
        self.logger.debug('Saying hello to {name}'.format(name=name))
        self.send('How are you {name}?!'.format(name=name), data)

    def handle_private(self, data, match):
        self.send('Hello, private message sender', data)


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

### More Examples

There is a repository with some services on my Github called
[tenyks-contrib](https://github.com/kyleterry/tenyks-contrib). These are all
using the older tenyksclient class and will probably work out of the box with
Tenyks. I'm going to work on moving them to the newer
[tenyks-service](https://github.com/kyleterry/tenyks-service) class.

A good example of something more dynamic is the [Weather
service](https://github.com/kyleterry/tenyks-contrib/blob/master/src/tenykswunderground/main.py).

## Help me

I'm a new Go programmer. Surely there's some shitty shit in here. You can help
me out by creating an issue on Github explaining how dumb I am. Or you can patch
the dumbness, make a pull request and tell me what I did wrong and why you made
the change you did. I'm open to criticism as long as it's done in a respectful
and "I'm teaching you something new" kind of way.

## Credit where credit is due

Service oriented anything isn't new. This idea came from an old
[coworker](https://github.com/wraithan) of mine. I just wanted to do something
slightly different. There are also plenty of other plugin style bots (such as
hubot and eggdrop). Every open source project needs love, so check those out as
well.
