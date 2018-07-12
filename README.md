[![CircleCI](https://circleci.com/gh/SYNQfm/SYNQ-Golang.svg?style=svg)](https://circleci.com/gh/SYNQfm/SYNQ-Golang)
[![Coverage Status](https://coveralls.io/repos/github/SYNQfm/SYNQ-Golang/badge.svg?branch=master)](https://coveralls.io/github/SYNQfm/SYNQ-Golang?branch=master)

## Introduction 

This is the Golang SDK for the SYNQ [API](https://docs.synq.fm)

## Installing
```
go get -u github.com/SYNQfm/SYNQ-Golang
```

## Usage

Here's an example of a simple main script that uses our SDK

```golang
package main

import (
  "log"

  "github.com/SYNQfm/SYNQ-Golang/synq"
)

func main() {
  // create API using username and password
  api := synq.Login("email", "password")
  // create API using a valid token
  api = synq.NewV2("token")
  video, _ := api.GetVideo("myvideo")
  log.Printf("video returned %v", video)
}
```

V2 Video [JSON](https://github.com/SYNQfm/SYNQ-Golang/blob/master/sample/video2.json)
```javascript
{
  "data": {

  }
}
```

### Utilizing the testing framework

There's a pretty powerful mocked server in test_server/server.go which can be used for testing your service connected to the SDK.  Here's an example of how to use it

```golang
```

## Usage (CLI)

You can also exercise the code via the command line using our `cli`.  View our more detailed [readme](https://github.com/SYNQfm/SYNQ-Golang/blob/master/cli/README.md)
