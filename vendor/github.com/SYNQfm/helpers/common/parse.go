package common

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	DEFAULT_DB_URL = "postgres://circleci:circleci@localhost:5432/db_test?sslmode=disable"
)

func GetDB(def_url ...string) string {
	db_url := os.Getenv("DATABASE_URL")
	if db_url == "" && len(def_url) > 0 {
		db_url = def_url[0]
	}
	dbaddr := ParseDatabaseUrl(db_url)
	return dbaddr
}

// this parses the database url and returns it in the format sqlx.DB expects
func ParseDatabaseUrl(dbUrl string) string {
	if dbUrl == "" {
		return ""
	}
	u, e := url.Parse(dbUrl)
	if e != nil {
		log.Printf("Error parsing '%s' : %s\n", dbUrl, e.Error())
		return ""
	}
	str := fmt.Sprintf("host=%s port=%s dbname=%s",
		u.Hostname(), u.Port(), strings.Replace(u.Path, "/", "", -1))
	if u.User != nil && u.User.Username() != "" {
		pass, set := u.User.Password()
		str = str + " user=" + u.User.Username()
		if set {
			str = str + " password=" + pass
		}
	}
	ssl := u.Query().Get("sslmode")
	if ssl != "" {
		str = str + " sslmode=" + ssl
	}
	return str
}

func IsSSL(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

//checks if a string is a valid html color(but without the '#')
func IsColor(str string) bool {
	result, _ := regexp.MatchString("(^[0-9A-Fa-f]{3}([0-9A-Fa-f]{3})?$)", str)
	return result
}

func IsNumber(str string) bool {
	result, _ := regexp.MatchString("(^[0-9]+$)", str)
	return result
}

func ParseBool(key string, urlMap url.Values, defaultValue ...bool) bool {
	if urlMap[key] != nil {
		if urlMap[key][0] == "true" {
			return true
		} else if urlMap[key][0] == "false" {
			return false
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

func ParseColor(key string, urlMap url.Values, defaultValue ...string) string {
	if urlMap[key] != nil {
		if IsColor(urlMap[key][0]) {
			return "#" + urlMap[key][0]
		}
	}
	if len(defaultValue) > 0 {
		return "#" + defaultValue[0]
	}
	return ""
}

func ParseInt(key string, urlMap url.Values, defaultValue ...int) int {
	if urlMap[key] != nil {
		if IsNumber(urlMap[key][0]) {
			seek, _ := strconv.Atoi(urlMap[key][0])
			return seek
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func ParseString(key string, urlMap url.Values, defaultValue ...string) string {
	if urlMap[key] != nil {
		return urlMap[key][0]
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}
