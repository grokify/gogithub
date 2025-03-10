# GoGitHub

[![Build Status][build-status-svg]][build-status-url]
[![Lint Status][lint-status-svg]][lint-status-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![LOC][loc-svg]][loc-url]
[![License][license-svg]][license-url]

 [build-status-svg]: https://github.com/grokify/gogithub/actions/workflows/test.yaml/badge.svg?branch=master
 [build-status-url]: https://github.com/grokify/gogithub/actions/workflows/test.yaml
 [lint-status-svg]: https://github.com/grokify/gogithub/actions/workflows/lint.yaml/badge.svg?branch=master
 [lint-status-url]: https://github.com/grokify/gogithub/actions/workflows/lint.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/gogithub
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/gogithub
 [codeclimate-status-svg]: https://codeclimate.com/github/grokify/gogithub/badges/gpa.svg
 [codeclimate-status-url]: https://codeclimate.com/github/grokify/gogithub
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/gogithub
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/gogithub
 [loc-svg]: https://tokei.rs/b1/github/grokify/gogithub
 [loc-url]: https://github.com/grokify/gogithub
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/gogithub/blob/master/LICENSE

`gogithub` is a high-level module to interact with GitHub.

The initial purpose is to search for pull open pull requests in accounts so that reports can be
generated for remediation.

The inclusion of checks is being investigated:

https://docs.github.com/en/rest/guides/using-the-rest-api-to-interact-with-checks?apiVersion=2022-11-28