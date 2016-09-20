package main

import (
    "os"
    "strings"
    "regexp"
    "net/url"
    "net/http"
)

func GetEnv(env string, defaultValue string) string {
    value := os.Getenv(env)
    if value == "" {
        value = defaultValue
    }

    return value
}

var slugRe = regexp.MustCompile("[^a-z0-9]+")
func slug(s string) string {
    return strings.Trim(slugRe.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func SetCorsHeaders(res http.ResponseWriter, req *http.Request, cors string, methods string) {
    if cors == "*" && len(req.Header["Referer"]) != 0 {
        referer, err := url.Parse(req.Header["Referer"][0])
        if err == nil {
            cors = referer.Scheme + "://" + referer.Host
        }
    }

    if methods == "" {
        methods = "GET,POST,PUT,DELETE"
    }

    res.Header().Set("Access-Control-Allow-Origin", cors)
    res.Header().Set("Access-Control-Allow-Methods", methods)
    res.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func sliceContains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}
