## Test Server

This is a mock Akka (and legacy obaku) server.  It will start a server and handle requests and return JSON objects back.  However, there are some basics you will need to know in order to use it

## Setup

To use the test server, you will need to create a `sample` directory under your repo.  So, let's use the example below.

```
my_app/
  sample/
    upload.json
    v2/
      login.json
      upload.json
  main.go
```

Some of them don't all fall into the `v2` category, so you'll have to look at the `handleV1` and `handleV2` functions to see how its loaded

## Usage

```
import (
  "http"

  "github.com/SYNQfm/SYNQ-Golang/test_server"
)

# create a "v2" server (supports Akka)
server := test_server.ServerServer("v2")
url := server.GetUrl()

# get all videos
http.Get(url + "/v1/videos")

# Get a specific video, with global video ids used by the server
# For instance V2_VIDEO_ID, is the valid id that you would call the get/upload apis for
http.Get(url + "/v1/videos/"+test_server.V2_VIDEO_ID)

# get the req and body values
reqs, vals := server.GetReqs()

# this is an array of url.Values, which has a single "body" key
body := vals[0].Get("body")[0]
json.Unmarshal(body, &obj)

# Get the last intiated server's reqs, since typically you only create one server per test
reqs, vals := test_server.GetReqs()
```

There's also helper function under `SYNQ-Golang/helper/setup.go` that will set an api object using the test server, see the code below

```
func SetupForTestV2() synq.ApiV2 {
  server := test_server.SetupServer(SYNQ_VERSION)
  url := server.GetUrl()
  api := synq.NewV2(test_server.TEST_AUTH)
  api.SetUrl(url)
  return api
}
```