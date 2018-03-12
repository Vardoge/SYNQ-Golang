package common

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"os"
	"strings"
	"time"
)

type Ret struct {
	Label    string
	CountMap map[string]int
	Error    error
	Start    time.Time
	DurMap   map[string]time.Duration
	BytesMap map[string]int64
}

func NewRet(label string) Ret {
	return Ret{
		Label:    label,
		DurMap:   make(map[string]time.Duration),
		BytesMap: make(map[string]int64),
		CountMap: make(map[string]int),
		Error:    nil,
		Start:    time.Now(),
	}
}

func ParseType(type_ string) string {
	t := strings.ToLower(type_)
	switch t {
	case "error":
		t = "errored"
	case "skip":
		t = "skipped"
	case "ct":
		t = "count"
	}
	return t
}

func (r *Ret) AddBytes(bytes int64) {
	r.AddBytesFor("total", bytes)
}

func (r *Ret) AddBytesFor(key string, bytes int64) {
	t := ParseType(key)
	if _, ok := r.BytesMap[t]; !ok {
		r.BytesMap[t] = bytes
	} else {
		r.BytesMap[t] = r.BytesMap[t] + bytes
	}
}

func (r *Ret) AddDurFor(key string, dur time.Duration) {
	t := ParseType(key)
	if _, ok := r.DurMap[t]; !ok {
		r.DurMap[t] = dur
	} else {
		r.DurMap[t] = r.DurMap[t] + dur
	}
}

func (r *Ret) Add(type_ string) {
	r.AddFor(type_, 1)
}

func (r *Ret) AddFor(type_ string, ct int) {
	t := ParseType(type_)
	if _, ok := r.CountMap[t]; !ok {
		r.CountMap[t] = 0
	}
	r.CountMap[t] = r.CountMap[t] + ct
}

func (r *Ret) AddError(err error) bool {
	if err != nil {
		r.Error = err
		r.Add("errored")
		return true
	}
	return false
}

func (r *Ret) IsErrored() bool {
	return r.Value("total") == r.Value("errored")
}

func (r *Ret) Value(type_ string) int {
	t := ParseType(type_)
	c, ok := r.CountMap[t]
	if !ok {
		c = 0
	}
	return c
}

func (r *Ret) Bytes(type_ string) int64 {
	t := ParseType(type_)
	b, ok := r.BytesMap[t]
	if !ok {
		b = int64(0)
	}
	return b
}

func (r *Ret) Duration(type_ string) time.Duration {
	t := ParseType(type_)
	d, ok := r.DurMap[t]
	if !ok {
		d = time.Duration(0)
	}
	return d
}

func (r *Ret) Eq(type_ string, ct int) bool {
	return r.Value(type_) == ct
}

func (r *Ret) Gte(type_ string, ct int) bool {
	return r.Value(type_) >= ct
}

func (r *Ret) Gt(type_ string, ct int) bool {
	return r.Value(type_) > ct
}

func (r *Ret) Lte(type_ string, ct int) bool {
	return r.Value(type_) <= ct
}

func (r *Ret) Lt(type_ string, ct int) bool {
	return r.Value(type_) < ct
}

func (r *Ret) LimitReached(limit int) bool {
	return r.Gte("count", limit) || r.Gte("errored", limit)
}

func Label(dur time.Duration) string {
	if dur == time.Hour {
		return "hrs"
	} else if dur == time.Minute {
		return "mins"
	} else if dur == time.Second {
		return "sec"
	} else if dur == time.Millisecond {
		return "ms"
	} else {
		return "ns"
	}
}

func DurFromLabel(label string) time.Duration {
	if label == "hrs" {
		return time.Hour
	} else if label == "mins" {
		return time.Minute
	} else if label == "sec" {
		return time.Second
	} else if label == "ms" {
		return time.Millisecond
	} else {
		return time.Nanosecond
	}

}

func DurVal(dur time.Duration) (int, string) {
	var t time.Duration
	if dur >= 1000*time.Minute {
		t = time.Hour
	} else if dur >= 1000*time.Second {
		t = time.Minute
	} else if dur >= 10000*time.Millisecond {
		t = time.Second
	} else if dur >= 10000*time.Nanosecond {
		t = time.Millisecond
	} else {
		t = time.Nanosecond
	}
	taken := int(dur / t)
	return taken, Label(t)
}

// This will determine the right value to use
func (r *Ret) Taken(tDur ...time.Duration) (int, string) {
	dur := time.Since(r.Start)
	if len(tDur) > 0 {
		t := tDur[0]
		taken := int(dur / t)
		return taken, Label(t)
	} else {
		return DurVal(dur)
	}
}

func BytesVal(bytes int64) (int64, string) {
	if bytes == 0 {
		return 0, ""
	}
	meg := int64(1000 * 1000)
	if bytes < 50000 {
		return bytes, "B"
	} else if bytes < meg {
		return bytes / 1000, "KB"
	}
	b := bytes / meg
	label := "MB"
	if b > 500000 {
		b = b / (1000 * 1000)
		label = "TB"
	} else if b > 5000 {
		b = b / 1000
		label = "GB"
	}
	return b, label
}

func (r *Ret) Speed() string {
	if bytes, ok := r.BytesMap["total"]; ok {
		secs, _ := r.Taken(time.Second)
		b, l := BytesVal(bytes)
		speed := (float64(b*8) / float64(secs))
		return fmt.Sprintf("%d %s (speed %.2f %sps)", b, l, speed, l)
	} else {
		return ""
	}
}

func (r *Ret) String() string {
	ct := r.Value("ct")
	total := r.Value("total")
	str := "for " + r.Label
	if ct > 0 || total > 0 {
		str = str + fmt.Sprintf(", processed %d/%d", ct, total)
	}
	for k, v := range r.CountMap {
		if k == "count" || k == "total" {
			continue
		}
		str = str + fmt.Sprintf(", %s %d", k, v)
		bytes := r.Bytes(k)
		dur := r.Duration(k)
		bStr := ""
		dStr := ""
		if bytes > 0 {
			avg := bytes / int64(v)
			b, l := BytesVal(bytes)
			a, l2 := BytesVal(avg)
			bStr = fmt.Sprintf("%d %s, avg %d %s", b, l, a, l2)
		}
		if dur > 0 {
			d, l := DurVal(dur)
			avg := int(dur*time.Nanosecond) / v
			d2, l2 := DurVal(time.Duration(avg))
			dStr = fmt.Sprintf("duration %d %s, avg %d %s", d, l, d2, l2)
		}
		if bStr != "" && dStr != "" {
			str = str + " (" + bStr + ", " + dStr + ")"
		} else if bStr != "" {
			str = str + " ( " + bStr + ")"
		} else if dStr != "" {
			str = str + " ( " + dStr + ")"
		}
	}
	str = str + "\n"
	if r.Error != nil {
		str = str + r.GetErrorString()
	}
	s, l := r.Taken()
	speed := r.Speed()
	if speed != "" {
		speed = ", " + speed
	}
	str = str + fmt.Sprintf("took %d %s%s", s, l, speed)
	return str
}

func (r *Ret) GetErrorString() string {
	return fmt.Sprintf("Error occured : %s\n", r.Error.Error())
}

// return 32 bytes into 36 bytes
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func ConvertToUUIDFormat(uuid string) string {
	if len(uuid) == 36 && strings.Count(uuid, "-") == 4 {
		return uuid
	}
	if len(uuid) != 32 {
		log.Printf("invalid uuid %s\n", uuid)
		return uuid
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", uuid[0:8], uuid[8:12], uuid[12:16], uuid[16:20], uuid[20:])
}

func ConvertFromUUIDFormat(uuid string) string {
	return strings.Replace(uuid, "-", "", -1)
}

func ValidUUID(uuid string) bool {
	if uuid == "" {
		return false
	}
	if len(uuid) == 36 {
		sec := strings.Split(uuid, "-")
		if len(sec) != 5 {
			return false
		}
		return len(sec[0]) == 8 && len(sec[1]) == 4 && len(sec[2]) == 4 && len(sec[3]) == 4 && len(sec[4]) == 12
	} else if len(uuid) == 32 {
		if strings.Contains(uuid, "-") {
			return false
		}
		return true
	} else {
		return false
	}
}

// Get environment variable
func GetOsEnv(env string, mandatory bool, defaultValue ...string) (value string) {
	value = os.Getenv(env)
	if value == "" {
		if mandatory {
			log.Panicf(`Can't run without environment variable ${%s} set.`, env)
		} else if len(defaultValue) > 0 {
			value = defaultValue[0]
		}
	}
	return
}

func GetAwsSignature(message, secret string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// Multipart Upload
func GetMultipartSignature(headers, awsSecret string) []byte {
	infoMap := map[string]string{
		"signature": GetAwsSignature(headers, awsSecret),
	}

	signature, _ := json.Marshal(infoMap)
	return signature
}

func GetFileExtension(ctype string) string {
	ext := CtypeToExt(ctype)
	if ext == "" {
		return ext
	}
	return strings.Split(ext, ".")[1]
}

func CtypeToExt(ctype string) string {
	exts, err := mime.ExtensionsByType(ctype)
	if err != nil {
		return ""
	}
	if len(exts) > 0 {
		return exts[0]
	}
	switch ctype {
	case "application/ttml+xml":
		return ".ttml"
	case "application/x-subrip":
		return ".srt"
	case "application/xml":
		return ".xml"
	case "video/mp4":
		return ".mp4"
	}
	return ""
}

func ExtToCtype(ext string) string {
	ctype := mime.TypeByExtension(ext)
	if ctype != "" && !strings.Contains(ctype, "text/plain") {
		return ctype
	}
	switch ext {
	case ".ttml":
		return "application/ttml+xml"
	case ".srt":
		return "application/x-subrip"
	case ".mp4":
		return "video/mp4"
	case ".xml":
		return "application/xml"
	}
	return ""
}

func FindString(list []string, find string) int {
	for idx, item := range list {
		if item == find {
			return idx
		}
	}
	return -1
}
