package main

import (
    "time"
)

type Stats struct {
    Connections    int32
    Features       int
    Uptime         int32
}

var stats = Stats {
    Connections: 0,
    Features: 0,
    Uptime: 0,
}

var started = int32(time.Now().Unix())

func GetStats() Stats {
    now := int32(time.Now().Unix())
    stats.Uptime = now - started

    return stats
}

func StatsIncrementConnections() {
    stats.Connections++
}

func StatsSetFeature(features int) {
    stats.Features = features
}