package common

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"
)

const (
	UPLOADED_STATE = `video.state == "uploaded" && video.userdata.hydra && video.userdata.hydra.job_state == "completed"`
	HAS_DASH       = `video.userdata.hydra.outputs.dash && video.userdata.hydra.outputs.dash.metadata`
	HAS_HLS        = `video.userdata.hydra.outputs.hls`
)

type Cli struct {
	Command  string
	Timeout  time.Duration
	Sleep    time.Duration
	Simulate bool
	Limit    int
	CacheDir string
	Flag     *flag.FlagSet
	Args     map[string]interface{}
}

func NewCli() Cli {
	cli := Cli{
		Flag: flag.NewFlagSet(os.Args[0], flag.ExitOnError),
	}
	cli.Args = make(map[string]interface{})
	return cli
}

func (c *Cli) DefaultSetup(msg, def string) {
	c.String("command", def, msg)
	c.String("simulate", "true", "simulate the transaction")
	c.Int("timeout", 120, "timeout to use for API call, in seconds, defaults to 120")
	c.Int("limit", 10, "number of actions to run")
	c.String("cache_dir", "", "cache dir to use for saved values")
}

func (c Cli) Println(msg string) {
	c.Printf(msg + "\n")
}

func (c Cli) Printf(msg string, args ...interface{}) {
	if c.Simulate {
		msg = "(simulate) " + msg
	}
	log.Printf(msg, args...)
}

func (c Cli) GetCacheFile(name string) string {
	if c.CacheDir == "" {
		return ""
	}
	return c.CacheDir + "/" + name + ".json"
}

func (c *Cli) String(name, def, desc string) {
	c.Args[name] = c.Flag.String(name, def, desc)
}
func (c *Cli) Int(name string, def int, desc string) {
	c.Args[name] = c.Flag.Int(name, def, desc)
}

func (c *Cli) GetString(name string) string {
	if _, ok := c.Args[name]; !ok {
		return ""
	}
	return *c.Args[name].(*string)
}

func (c *Cli) GetInt(name string) int {
	if _, ok := c.Args[name]; !ok {
		return -1
	}
	return *c.Args[name].(*int)
}

func (c *Cli) GetBool(name string) bool {
	if _, ok := c.Args[name]; !ok {
		return false
	}
	return *c.Args[name].(*string) == "true"
}

func (c *Cli) GetSeconds(name string) time.Duration {
	val := c.GetInt(name)
	if val == -1 {
		return -1
	}
	return time.Duration(val) * time.Second
}

func (c *Cli) GetSleep() time.Duration {
	if c.Sleep != time.Duration(0) {
		return c.Sleep
	}
	return c.GetSeconds("sleep")
}

func (c *Cli) Parse(args ...[]string) {
	var a []string
	if len(args) > 0 {
		a = args[0]
	} else {
		i := 0
		for idx, val := range os.Args {
			if idx == 0 || strings.Contains(val, "-test.") {
				i = i + 1
				continue
			} else {
				break
			}
		}
		a = os.Args[i:]
	}
	c.Flag.Parse(a)
	c.Command = c.GetString("command")
	c.Timeout = c.GetSeconds("timeout")
	c.Simulate = c.GetString("simulate") != "false"
	c.Limit = c.GetInt("limit")
	c.CacheDir = c.GetString("cache_dir")
	c.Sleep = c.GetSeconds("sleep")
}

func (c *Cli) GetFilter(fta ...string) (filter string, filterType string) {
	var ft string
	if len(fta) >= 0 {
		ft = fta[0]
	} else {
		ft = c.GetString("filter_type")
	}
	checks := []string{"video.userdata.importer"}
	projectId := c.GetString("project")
	hasProject := `video.userdata["` + projectId + `"] != null`
	switch ft {
	case "metadata_only",
		"mo":
		checks = append(checks, `video.state == "created"`)
		filterType = "mo"
	case "all_with_metadata",
		"awm":
		checks = append(checks, hasProject)
		filterType = "awm"
	case "no_metadata",
		"nm":
		checks = append(checks, "video.userdata.juneMeta == null")
		filterType = "nm"
	case "dash_only",
		"do":
		checks = append(checks, UPLOADED_STATE)
		checks = append(checks, HAS_DASH)
		checks = append(checks, `video.userdata.hydra.outputs.hls == null`)
		filterType = "do"
	case "hls_dash",
		"hlda":
		checks = append(checks, UPLOADED_STATE)
		checks = append(checks, HAS_DASH)
		checks = append(checks, HAS_HLS)
		filterType = "hlda"
	case "transcoded",
		"t":
		checks = append(checks, UPLOADED_STATE)
		filterType = "t"
	case "series_dash",
		"sd":
		checks = append(checks, UPLOADED_STATE)
		checks = append(checks, HAS_DASH)
		checks = append(checks, hasProject)
		checks = append(checks, `video.userdata["`+projectId+`"]["seriesInfo"] != null`)
		checks = append(checks, `video.userdata["`+projectId+`"]["seriesInfo"]["seriesId"] == "`+c.GetString("series_id")+`"`)
		filterType = "sd"
	case "all_duration",
		"ad":
		checks = append(checks, UPLOADED_STATE)
		checks = append(checks, HAS_DASH)
		filterType = "ad"
	default:
		log.Printf("unknown type '%s'\n", ft)
		os.Exit(1)
	}
	toFilter := strings.Join(checks, " && ")
	filter = `if (` + toFilter + `) { return video }`
	return filter, filterType
}
