# Morning Pages

A [Morning Pages](http://juliacameronlive.com/basic-tools/morning-pages/) app written in Go for [scriptio continua](http://en.wikipedia.org/wiki/Scriptio_continua).

[![Build Status](https://travis-ci.org/shuhei/morning_pages.png)](https://travis-ci.org/shuhei/morning_pages)
[![Coverage Status](https://coveralls.io/repos/shuhei/morning_pages/badge.png?branch=HEAD)](https://coveralls.io/r/shuhei/morning_pages?branch=HEAD)

## Installation

1. Install Go.
2. Make sure that `$GOPATH/bin` is in your `$PATH`.
3. Install godep. `go get github.com/kr/godep`
4. Pull this repository.
5. Set environmental variables. In development, use `.env`.
6. `godep go install` and `foreman start`

## Environmental Variables

- `MARTINI_ENV` : `development` or `production`
- `MONGOHQ_URL` : MongoDB URL
- `FB_APP_ID` : Facebook app ID
- `FB_APP_SECRET` : Facebook app secret
- `FB_REDIRECT_URL` : Facebook redirect URL
- `SESSION_KEY` : secret session key

## Test

```
godep go test
```

## Add or update a dependency

This app's dependencies are managed with [godep](https://github.com/kr/godep). To add or update a dependency, see [its README](https://github.com/kr/godep).

