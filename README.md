# vrka
Vrka - is a distributed callback engine. Clients can set a callback to be invoked after specific time interval. Vrka then invokes this callback after the time interval is over.

Work in progress.

Vrka - a distributed callback engine in GO-LANG
[Design details]
What's Vrka:
> The client can register a callback (a HTTP URI) to be invoked after a given time interval
> At the scheduled time, Vrka invokes the callback

Desirable properties:
> Fault tolerant by design
> Use commodity hardware
> Horizontally scalable 

A simple interface:
> Add [time interval in ms, callback to uri]
> The client passes the message “Add” to the server with parameters:
    time interval (in milliseconds) after which the “uri” has to be invoked
    To begin with Vrka will only support a HTTP uri
> The client and server will communicate using a simple text based protocol
   The details of the protocol are not discussed in the presentation
  The code provides a sample client to show the use of the protocol

How do we represent "time"?
> Vrka has to maintain the callbacks to be invoked, and invoke them after the scheduled “time” has reached
> Do use wall clock time?
   Cannot use wall clock of the system to represent time
> Wall clock time can move back or forward if the system time is reset
> Use a counter that keeps decrementing?
   No
   No reliable way to decrement counter periodically - the system could freeze up, go down
> Viable alternative - a monotonic counter




