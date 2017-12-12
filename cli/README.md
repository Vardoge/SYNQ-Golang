## Command Line Interface

This can handle both versions of our API.

## V2 Examples

```
# Upload File for specific asset
./cli -command=upload -version=v2 -api_key=<token> \
 -file=$1 -simulate=$2 \
 -asset_id=99cfb5d7-29c5-4f7f-8e56-074895b1707a \
 -cache_dir=cache_dir
```

## V1 Examples

```bash
# Create a new video object
./cli -api_key=<key> -command create -version=v1
# Upload a file
./cli -api_key=<key> -video_id=<vid> -file <file name> -command upload -version=v1
# Get details for a video
./cli -api_key=<key> -video_id=<vid> -command details -version=v1
```


### General Usage
```
cd cli
go build
./cli -h

Usage of ./cli:
Usage of ./cli:
  -api_key string
    	pass the synq api key
  -asset_id string
    	asset id to access
  -cache_dir string
    	cache dir to use for saved values
  -command string
    	upload (default "for v2 'upload', get_video', for v1 : details, upload_info, upload, create, uploader_info, uploader, query or create_and_then_multipart_upload")
  -file string
    	path to file you want to upload or userdata
  -limit int
    	number of actions to run (default 10)
  -password string
    	password to use
  -query string
    	query string to use
  -simulate string
    	simulate the transaction (default "true")
  -timeout int
    	timeout to use for API call, in seconds, defaults to 120 (default 120)
  -upload_url string
    	upload url to use (default "http://s6krcbatzuuhmspse.stoplight-proxy.io")
  -user string
    	user to use
  -version string
    	version to use (default "v2")
  -video_id string
    	video id to access
```
