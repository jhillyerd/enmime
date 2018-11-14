# enmime
[![GoDoc](https://godoc.org/github.com/jhillyerd/enmime?status.png)][GoDoc]
[![Build Status](https://travis-ci.org/jhillyerd/enmime.png?branch=master)][Build Status]
[![Go Report Card](https://goreportcard.com/badge/github.com/jhillyerd/enmime)][Go Report Card]
[![Coverage Status](https://coveralls.io/repos/github/jhillyerd/enmime/badge.svg)][Coverage Status]

enmime is a MIME encoding and decoding library for Go, focused on generating and
parsing MIME encoded emails.  It is being developed in tandem with the
[Inbucket] email service.

enmime includes a fluent interface builder for generating MIME encoded messages,
see the wiki for example [Builder Usage].

API documentation and examples can be found here:
http://godoc.org/github.com/jhillyerd/enmime


## API Changes

Part readers: `Part.Read()` and `Part.Utf8Reader` have been removed. Please use
`Part.Content` instead.

A brief guide to migrating from the old 2016 go.enmime API is available here:
https://github.com/jhillyerd/enmime/wiki/Enmime-Migration-Guide


## Development Status

enmime is beta quality: it works but has not been tested with a wide variety of
source data.  It's possible the API will evolve slightly before an official
release.

Please see [CONTRIBUTING.md] if you'd like to contribute code to the project.


## About

enmime is written in [Google Go][Golang].

enmime is open source software released under the MIT License.  The latest
version can be found at https://github.com/jhillyerd/enmime

[Build Status]:    https://travis-ci.org/jhillyerd/enmime
[Builder Usage]:   https://github.com/jhillyerd/enmime/wiki/Builder-Usage 
[Coverage Status]: https://coveralls.io/github/jhillyerd/enmime
[CONTRIBUTING.md]: https://github.com/jhillyerd/enmime/blob/develop/CONTRIBUTING.md
[Inbucket]:        http://www.inbucket.org/
[GoDoc]:           https://godoc.org/github.com/jhillyerd/enmime
[Golang]:          http://golang.org/
[Go Report Card]:  https://goreportcard.com/report/github.com/jhillyerd/enmime
