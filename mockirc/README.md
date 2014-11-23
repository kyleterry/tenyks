# MockIRC

This is a mock IRC server for testing go programs that interact with IRC. This
is NOT a real IRC server... So don't attempt to use it as one.

## Usage

```go
package test

import (
  "testing"
  "github.com/kyleterry/tenyks/mockirc"
)

func TestMyIRCInteraction(t *testing.T) {
  ircServer := MockIRC.New("mockirc.tenyks.io", 6661) // servername and port
  // When I recieve "PING mockirc.tenyks.io" on the server, respond back with PONG...
  ircServer.When("PING mockirc.tenyks.io").Respond(":PONG mockirc.tenyks.io")
  ircServer.When("NICK kyle").Respond("... response to NICK")
  wait, err := ircServer.Start()
  if err != nil {
    t.Error("Error starting mockirc")
  }
  defer ircServer.Stop()
  <-wait // wait for start to fire up channel

  myircthing.Dial("localhost:6661")
  myircthing.Bootstrap() //send irc things and what not

  if !myircthing.Connected {
    t.Error("my shit must be broken because I didn't connect to the flawless mockirc")
  }
}
```
