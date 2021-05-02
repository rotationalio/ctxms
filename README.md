# Contextual Microservices

**Example code for the blog post "[Contexts in Go Microservice Chains](https://rotational.io/blog/contexts-in-go-microservice-chains/)"**

## Running the Experiment

After cloning the repository, open 6 different terminals for each of the services and a seventh to run the client code. Map each of the following commands to one of the server terminals:

    $ go run ./cmd/ctxms serve -n alpha -p 9000 -d 10s
    $ go run ./cmd/ctxms serve -n bravo -p 9001 -d 3s
    $ go run ./cmd/ctxms serve -n charlie -p 9002 -d 7s
    $ go run ./cmd/ctxms serve -n delta -p 9003 -d 5s
    $ go run ./cmd/ctxms serve -n echo -p 9004 -d 8s
    $ go run ./cmd/ctxms serve -n foxtrot -p 9005 -d 4s -t

Note that it is very important to start at port 9000 and to increment the port numbers by 1. Also, note the unique names and the `-t` flag on the last server - this ensures the service forms a ring where one microservice will pass a message to the next one in a circle. The `-d` flag specifies the maximum amount of a random delay for "hard work".

You can then run a client command as follows:

    $ go run ./cmd/ctxms trace -e localhost:9003 -t 12s

This will run a trace to the "delta" server with a timeout of 12 seconds. Experiment with different timeouts to see the context deadline exceeding at different servers.
