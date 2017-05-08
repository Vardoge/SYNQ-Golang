[![CircleCI](https://circleci.com/gh/SYNQfm/SYNQ-Golang.svg?style=svg)](https://circleci.com/gh/SYNQfm/SYNQ-Golang)
[![Coverage Status](https://coveralls.io/repos/github/SYNQfm/SYNQ-Golang/badge.svg?branch=master)](https://coveralls.io/github/SYNQfm/SYNQ-Golang?branch=master)

## Introduction 

This is the Golang SDK for the SYNQ [API](https://synq.fm/docs)

## Usage

Here's an example of a simple main script that uses our SDK

```golang
package main

import (
  "log"

  "github.com/SYNQfm/SYNQ-Golang/synq"
)

func main() {
  api := synq.New("myapikey")
  video := api.GetVideo("myvideo")
  log.Printf("video returned %v", video)
}
```