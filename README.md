# Portal

Scalable patterns for concurrent software

[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/lthibault/portal)

- [Portal](#portal)
    - [Overview](#overview)
    - [Thinking with portals](#thinking-with-portals)
        - [Libraries, not frameworks](#libraries-not-frameworks)
        - [Principle of least surprise](#principle-of-least-surprise)
        - [Socket-like](#socket-like)
        - [Channel-like](#channel-like)
    - [Bare-bones example](#bare-bones-example)
    - [Patterns](#patterns)
        - [Type Guards](#type-guards)
        - [Supervision Trees](#supervision-trees)
    - [Authors](#authors)
    - [License](#license)
    - [Acknowledgments](#acknowledgments)
    - [Request for Comments](#request-for-comments)


## Overview


Portal is a lightweight library that provides concurrent communication patterns.  It's primary goals are to:

1. make it ~~easy~~ easier to reason about concurrent code
2. reduce the amount of boilerplate involved in managing concurrency
3. encorage modular, scalable software design

While Go provides concurrency primitives as part of the core language, it's easy to spend more time managing concurrency than writing the application logic.  Just as one avoids writing HTTP servers from raw sockets, one would do well to avoid writing concurrent software from raw goroutines, channels and mutices.

The solution is to use portals.

Portals are channels that implement in-process messaging patterns, called _protocols_.
Portals allow safe communication across goroutines, and easy management of concurrent data-flow.
Portal provides basic protocols, which can be combined to form more advanced topologies:

1. **Pair:**  One-to-one, bidirectional communication
1. **Bus:**  Many-to-many (broadcast) communication
1. **Request / Reply:**  "I ask, you answer".  Process requests & return a response
1. **Publish / Subscribe:**  One-to-many distribution to interested subscribers
1. **Push / Pull:**  Pipeline pattern (unidirectional data flow)
1. **Surveyor / Respondent:**  Query multiple components, each of which can reply

These protocols behave similarly to their [nanomsg](http://nanomsg.org/gettingstarted/index.html) counterparts.  (_N.B._:  use nanomsg if you need portal-like behavior over the network.)

Portal can be installed with the standard go toolchain:

```
go get -u github.com/lthibault/portal
```

## Thinking with portals

### Libraries, not frameworks

Portal doesn't force you to think about your program in any particular way.
Instead, it provides you with building blocks that help manage concurrency while leaving the architectural design up to you.  This approach is considered idiomatic in Go, and is most famously embodied in `net/http` to great effect.

### Principle of least surprise

Portal strives to be consistent with Go's channel API.

1. Sening to a closed portal panics
1. Reveiving from a closed portal returns `nil`
1. Portals can be unbuffered (synchronous) or buffered (asynchronous)

### Socket-like

Like sockets, portals communicate with each other by _binding_ a portal to an address, after which other sockets can _connect_ to that same address.

Portal addresses are human-readable strings.  The only constraint is that only one portal can `Bind` to an address; a common convention is to use `/`-separated paths as addresses.  For example:  `/stream/input`.  See [below](#bare-bones-example) for an example.

### Channel-like

The `Portal` interface is characterized by two methods in particular:

```go
type interface Portal {
    // ... omitted for clarity

    Send(interface{})
    Recv() interface{}
}
```

These implement channel-like semantics that allow applications developers to _send_ and _receive_ data safely across goroutines.  The data written through one portal by a call to `Send` can be read by a connected portal via a call to `Recv`.

By default, portals are unbuffered and synchronous.  This means that subsequent calls to `Send` will block until **all** connected portals have called `Recv`.  With buffered (asynchronous) portals, subsequent calls to `Send` will not block until the buffer is full.  **However**, subsequent calls to `Recv` on a given portal will block until all other connected portals have called `Recv`.

## Bare-bones example

Using portals is a very simple process that will feel very familiar:

1. Import the protocols you plan to use.
1. Instantiate the necessary `Portal` instances using the `New` function in each protocol submodule.
1. `Bind` a portal to an address
1. `Connect` to the above portal at its bound address
1. Call `Send`/`Recv` as needed

```go
/* A bare-bones Push-Pull example */

import (
    "log"

    // step 1
    "github.com/lthibault/portal"
    "github.com/lthibault/portal/proto/push"  
    "github.com/lthibault/portal/proto/pull"  
)

// datum is some arbitrary thing we want to move around
type datum struct {
    Foo string
    Bar int
}

func main() {
    // Step 2
    consumer := pull.New(portal.Cfg{}) // synchronous by default
    producer := push.New(portal.Cfg{})  

    // Step 3
    if err := consumer.Bind("/some/arbitrary/addr"); err != nil {
        panic(err)
    }
    defer consumer.Close()

    // Step 4
    if err := producer.Connect("/some/arbitrary/addr"); err != nil {
        panic(err)
    }
    defer producer.Close()

    // Step 5
    go func() {
        for {
            log.Println(consumer.Recv())   
        }
    }()

    for _, d := range []datum{{Foo: "hello"}, {Foo: "world", Bar: 1}} {
        producer.Send(d)
    }
}
```

## Patterns

This section illustrates idiomatic design patterns for Portal.  These will help you structure your applications and avoid pitfalls.

### Type Guards

`interface{}`... `interface{}` everywhere...

Portal relies on `interface{}` to enable the sending of arbitrary types through
a `Portal` instance.  As such, it is recommended that application developers wrap
`Portal` instances in a lightweight structure that overrides `Send` and `Recv` methods to enforce type-safety.

We call this the  *type guard* pattern.

```go
type fooGuard struct { Portal }

func (fg fooGuard) Send(f Foo) { fg.Portal.Send(f) }
func (fg fooGuard) Recv() Foo { return fg.Portal.Recv().(Foo) }
```

### Supervision Trees

A typical Portal application will contain several (if not dozens) of `Portal` instances.  [Supervision trees](http://www.jerf.org/iri/post/2930) are a good way handling start-up and shut-down logic in such applications.

If you're unfamiliar with supervision trees, we highly suggest you familiarize yourself with;

1. [The Zen of Erlang](https://ferd.ca/the-zen-of-erlang.html) (a fantastic read!)
1. Jeremy Bowers' [Suture](http://github.com/thejerf/suture) library

For those familiar with the basic concept, we draw your attention to the fact that this pattern is idiomatic go insofar as it is an example of [_programming_ with errors](https://blog.golang.org/errors-are-values).

By implementing `suture.Service`, we can ensure greaceful retry/restart/backoff behavior when connecting portals, and recover from any errors.  As a bonus, our application becomes extremely resilient to failure.

```go
type PairBinder struct {
    portal.Portal
    addr string
}

// Serve implements suture.Service
func (p *PairBinder) Serve() {
    p.Portal = pair.New(portal.Cfg{})

    if err := p.Bind(p.addr); err != nil {
        log.Println(errors.Wrap(err, "PairBinder failed to Bind"))
        return  // suture.Supervisor will handle restart/backoff logic
    }

    for {
        // Application logic goes here...
        // If an error occurs, log & return
    }
}

func (p PairBinder) Stop() { p.Close() }
```

An analogous `PairConnector` would then handle the connecting `Portal` instance(s).  Further generalizations of this pattern are possible, but the scope of this tutorial.

## Authors

* **Louis Thibault** - *Initial work* - [lthibault](https://github.com/lthibault)

<!-- See also the list of [contributors](https://github.com/your/project/contributors) who participated in this project. -->

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgments

The following projects significantly influenced the design of Portal.

* [mangos](https://github.com/go-mangos/mangos)

## Request for Comments

If you use Portal in any of your projects -- professional or otherwise -- please
drop me a line at `l.thibault@sentimens.com` to let me know:

1. what you're building
2. your thoughts on Portal
