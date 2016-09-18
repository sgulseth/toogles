package main

import (
    "os"
    "strings"
    "regexp"
)

func GetEnv(env string, defaultValue string) string {
    value := os.Getenv("PORT")
    if value == "" {
        value = defaultValue
    }

    return value
}

var slugRe = regexp.MustCompile("[^a-z0-9]+")
func slug(s string) string {
    return strings.Trim(slugRe.ReplaceAllString(strings.ToLower(s), "-"), "-")
}