```Textile
  _                   _         
 | |                 | |        
 | |_ ___ _ __  _   _| | _____  
 | __/ _ \ '_ \| | | | |/ / __| 
 | ||  __/ | | | |_| |   <\__ \  irc bot.
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
[here](./legacy/tenyks/client).

## Installation and building

Since Tenyks is pretty new, you will need to build the bot yourself. Step 1 is
making sure you have a redis-server running. I won't go into detail there as I'm
sure you can figure it out. Step 2 is making sure you have all the Go
dependencies installed. For instance on Debian you would run `sudo apt-get
install golang`.

### Building

You can build with the [Go tool](http://golang.org/cmd/go/).

```bash
git clone https://github.com/kyleterry/tenyks
cd tenyks
go build
go install
```

### Configuration

Configuration is just json. The included example contains everything you need to
get started. You just need to swap out the server information.

```bash
cp config.json.example ${HOME}/tenyks-config.json
```

## Running

`tenyks ${HOME}/tenyks-config.json`

## Testing

I'm a horrible person. There aren't tests yet. I'll get right on this...

## Services

TODO

## Help me

I'm a new Go programmer. Surely there's some shitty shit in here. You can help
my out by creating an issue on Github explaining how dumb I am. Or you can patch
the dumbness, make a pull request and tell me what I did wrong and why you made
the change you did. I'm open to criticism as long as it's done in a respectful
and "I'm teaching you something new" kind of way.

## Credit where credit is due

Service oriented anything isn't new. This idea came from an old
[coworker](https://github.com/wraithan) of mine. I just wanted to do something
slightly different. There are also plenty of other plugin style bots (such as
hubot and eggdrop). Every open source project needs love, so check those out as
well.
