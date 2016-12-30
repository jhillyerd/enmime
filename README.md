enmime
================================================================================

[![GoDoc](https://godoc.org/github.com/jhillyerd/enmime?status.png)][GoDoc]
[![Build Status](https://travis-ci.org/jhillyerd/enmime.png?branch=master)][Build Status]
[![Go Report Card](https://goreportcard.com/badge/github.com/jhillyerd/enmime)][Go Report Card]
[![Coverage Status](https://coveralls.io/repos/github/jhillyerd/enmime/badge.svg)][Coverage Status]

enmime is a MIME parsing library for Go.  It's built on top of Go's included
mime/multipart support, but is geared towards parsing MIME encoded emails.

It is being developed in tandem with the [Inbucket] email service.

API documentation can be found here:
http://godoc.org/github.com/jhillyerd/enmime

A brief guide to migrating from the old go.enmime API is available here:
https://github.com/jhillyerd/enmime/wiki/Enmime-Migration-Guide


## Development Status

enmime is approaching beta quality: it works but has not been tested with a wide
variety of source data.  It's possible the API will evolve slightly before an
official release.

Please see [CONTRIBUTING.md] if you'd like to contribute code to the project.


## About

enmime is written in [Google Go][Golang].

enmime is open source software released under the MIT License.  The latest
version can be found at https://github.com/jhillyerd/enmime

[Build Status]:    https://travis-ci.org/jhillyerd/enmime
[Coverage Status]: https://coveralls.io/github/jhillyerd/enmime
[CONTRIBUTING.md]: https://github.com/jhillyerd/enmime/blob/develop/CONTRIBUTING.md
[Inbucket]:        http://www.inbucket.org/
[GoDoc]:           https://godoc.org/github.com/jhillyerd/enmime
[Golang]:          http://golang.org/
[Go Report Card]:  https://goreportcard.com/report/github.com/jhillyerd/enmime
