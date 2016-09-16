package main

import (
    "os"
)

func GetEnv(env string, defaultValue string) string {
    value := os.Getenv("PORT")
    if value == "" {
        value = defaultValue
    }

    return value
}