What
====

Meetup is a micro proxy program that enables to connect two endpoints via TCP sockets.
Either of endpoint may listen or connect.

Disconnect on one end of the pipe breaks the connection to the other end.
`-connect` mode forever attempts to reconnect with 5 second interval between attempts.


Install
=======

`go get -u github.com/temoto/meetup`


Example
=======

 - you have a faulty PHP application and you want to debug it with xdebug
 - xdebug can connect to your machine:9000, but you are behind NAT
 - so you run `meetup -listen1=:9000 -listen2=:9001` on the application server
 - and another `meetup -connect1=appserver:9001 -connect=localhost:9000` on your machine
 First instance listens two ports and when a connection arrives on both, it creates
 a bidirectional buffered pipe between the two. The other instance connects to
 first meetup on appserver:9001 and also to your local IDE :9000 and likewise
 pipes data in both directions.
