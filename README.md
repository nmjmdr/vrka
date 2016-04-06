Vrka 
===================
Vrka - is a distributed callback engine. Clients can set a callback to be invoked after specific time interval. Vrka then invokes this callback after the time interval is over.

----------


Design
-------------

**Interface | Protocol**

Add [time interval in ms, callback to uri]

The client passes the message “Add” to the server with parameters:  "time interval" (in milliseconds) and the “uri” has to be invoked after the time interval expires.

> Note: To begin with Vrka will only support a HTTP uri


 The client and server will communicate using a simple text based protocol. Code provides a sample client to show the use of the protocol.


**How do we represent “time”**

Vrka has to maintain the callbacks to be invoked, and invoke them after the scheduled “time” has reached. We need way to represent time. We cannot use wall clock of the system to represent time as the wall clock time can move back or forward if the system time is reset.  We also cannot use a counter that keeps getting decremented As there is no reliable way to decrement counter periodically - the system could freeze up or go down.


A viable alternative is to use a monotonic counter (In GO: UnixNano())




	 
