# 0.10.0

* Added backoff to failed connections when they try to reconnect. This replaces
  the retries setting that will stop retrying when the max is reached.
* Kind of fixed a bug in the connection watchdog that would try to send a PING
  on a nil channel when a connection was disconnected.
* Fixed logging. It was terrible. It's still not great, but it's much better.
