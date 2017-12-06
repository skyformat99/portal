# Portal

Scalable patterns for concurrent software

[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/lthibault/portal)

## Overview

Portal is a lightweight library that provides concurrent communication patterns.  It's primary goals are to:

1. make it ~~easy~~ easier to reason about concurrent code
2. reduce the amount of boilerplate involved in managing concurrency
3. encorage modular, scalable software design

While Go provides concurrency primitives as part of the core language, it's easy to spend more time managing concurrency than writing the application logic.  Just as one avoids writing HTTP servers from raw sockets, one would do well to avoid writing concurrent software from raw goroutines, channels and mutices.

The solution is to use `Portal`s.

A portal is a channel analog that implements in-process messaging patterns such as *Request-Reply*, *Publish-Subscribe*, *Surveyor-Respondent*, *Pair*, *Bus*, *Push-Pull*, etc.  Portals are always thread-safe and syncrhonous by default, though they can be configured as async (buffered) portals as well.

### a library, not a framework

Portal doesn't force you to think about your program in any particular way.
Instead, it provides you with building blocks that help manage concurrency while leaving the architectural design up to you.  This approach is considered idiomatic in Go, and is most famously embodied in `net/http` to great effect.

### simple by default, fast on demand

Portal adopts the principle of least surprise.  When in doubt, Portal tries to be consistent with Go's channel API.  For example:

1. `Send` to a closed `Portal` panics
1. `Recv` from a closed `Portal` returns `nil`
1. A `Portal` is synchronous (unbuffered) by default

Portals are configured via the `portal.Cfg` struct.

### Portal semantics, buffering & concurrency

Portals are a bit like sockets in that they communicate by `Bind`ing and `Connect`ing to an address.  Some protocols define one-to-many or many-to-many patterns of messaging, which presents a few subtleties relative to `chan`'s behavior.

First and foremost, it **all** receiving portals **MUST** call `Recv` before any receiver can receive the next item.  This head-of-line blocking is necessary to ensure that messages are never dropped from a Portal topology.

`Send` operations on synchronous portals will block until _all_ connected Portals have called `Recv`.  As such, all `Recv` calls will unblock at the same time, ensuring that all receivers are in lock-step.  `Send` operations on asynchronous portals will not block unless the `Portal` is at maximum capacity.  Likewise, `Recv` calls will immediately return the first item in the buffer.

## Getting Started

This section is divided into three parts:

1. Installation:  get you a copy of Portal on your local machine
1. Bare-bones example:  get introduced to Portal's primitives & work-flow
1. Patterns:  idiomatic patterns that help you go from "I know the syntax" to "I know how to start"

Portal has been tested on Go 1.9.2 and above.

### Installation

```
go get -u github.com/lthibault/portal
```

### Bare-bones example

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
    "github.com/lthibault/portal/protocol/push"  
    "github.com/lthibault/portal/protocol/pull"  
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

### Patterns

This section illustrates idiomatic design patterns for Portal.  These will help you structure your applications and avoid pitfalls.

#### Type Guards

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

#### Supervision Trees

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

## Further reading

Documentation an examples for individual protocols are located in the `README.md` in each protocol subpackage (e.g.:  `protocols/sub/README.md`).

If anything is unclear, please file an issue with a bit of context, and we'll make sure the documentation is improved.

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
