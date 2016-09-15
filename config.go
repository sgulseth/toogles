package main

import (
    "log"
    "net/http"

    "encoding/json"
)


type Feature struct {
    Name            string
    Persistent      bool
    ShareStrategy   *ShareStrategy `json:",omitempty"`
    FirstStrategy   *FirstStrategy `json:",omitempty"`
    QueryStrategy   *QueryStrategy `json:",omitempty"`
}

func (self *Feature) Toggle(req *http.Request) bool {
    if self.ShareStrategy != nil {
        return self.ShareStrategy.Toggle(req)
    } else if self.FirstStrategy != nil {
        return self.FirstStrategy.Toggle(req)
    } else if self.QueryStrategy != nil {
        return self.QueryStrategy.Toggle(req)
    }

    return false
}

type Configuration struct {
    Features    []Feature
}

type FeatureEnvelope struct {
    StrategyType string
    Msg          interface{}
}

var globalConfig Configuration;

func setConfig(config Configuration) {
    globalConfig = config

    StatsSetFeature(len(globalConfig.Features))
}

func setConfigString(conf string) {
    var config Configuration
    if err := json.Unmarshal([]byte(conf), &config); err != nil {
        log.Fatal(err)
    }

    setConfig(config)
}

func getConfig() Configuration {
    return globalConfig
}

func saveConfigToRedis() {
    configBytes, _ := json.Marshal(globalConfig)
    configJson := string(configBytes[:])
    log.Print("Saving configuration to redis")
    log.Print(configJson)

    client := getRedisClient()
    err := client.Set("toogles_config", configJson, 0).Err()

    if err != nil {
        log.Fatal(err)
    }
}

func loadConfigFromRedis() {
    client := getRedisClient()
    data, err := client.Get("toogles_config").Result()

    var config Configuration
    if err = json.Unmarshal([]byte(data), &config); err != nil {
        log.Print("Could not find configurtaion in redis")
        return
    }

    setConfig(config)

    log.Print("Loaded configuration from redis")
    log.Print(data)
}
