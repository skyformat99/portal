# Portal

Scalable patterns for concurrent software

[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/lthibault/portal)

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

Portal has been tested on Go 1.9.2, but is expected to run on all versions above 1.6.

### Installation

```
go get -u github.com/lthibault/portal
```

### Development

## Running tests

Tests are located in the `test` directory.

## Running the benchmarks

Bencharks exist for each protocol, and are found in the `benchmark` directory.
Benchmarks can be run for a given protocol with the following command:

```
go test -bench=. benchmark/<proto>/<proto>_test.go
```

## Contributing

Contributor guidelines are intentionally brief and flexible.  Please do the
following prior to submitting a pull-request.

1. Open an issue stating a problem or potential improvement
2. State your interest in contributing a solution to the above issue

These steps are just to ensure that you don't end up spending a lot of time on a
PR that will be rejected :)

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags). 

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
