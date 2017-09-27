# wrk-go [![Build Status](https://travis-ci.org/arthurkiller/wrk-go.svg?branch=master)](https://travis-ci.org/arthurkiller/wrk-go) [![Go Report Card](https://goreportcard.com/badge/github.com/arthurkiller/wrk-go)](https://goreportcard.com/report/github.com/arthurkiller/wrk-go)
A wrk written in go

## What is wrk
wrk is a https testing tool util written in c. It is really easy to use but has limitted support with
https (ssl) extension.

## Why wrk-go
We need a tool with full suport with ssl extension feature. It should be easy to use and powerfull (like wrk), easy to
extend (with go).

## Feature

* made in go with ❤️
* support ssl status with OSCP stapling
* support NPN/APLN, ssl false start support
* support session ticket for session resumption
* histogram supported
