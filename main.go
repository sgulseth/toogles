package main

import (
    "os"
    "fmt"
    "time"
    "strings"
    "bytes"
    "net/http"
    "log"
    "io/ioutil"
    "encoding/json"
)

var Port = GetEnv("PORT", "8080")

func HandlerNotFound(res http.ResponseWriter, req *http.Request) {
    res.WriteHeader(http.StatusNotFound)
    fmt.Fprint(res, "404 page not found")
}

func HandleToggles(res http.ResponseWriter, req *http.Request) {
    defer req.Body.Close()

    if req.URL.Path != "/" {
        HandlerNotFound(res, req)
        return
    }

    StatsIncrementConnections()

    res.Header().Set("Connection", "Close")
    res.Header().Set("Content-Type", "application/json")
    res.Header().Set("Access-Control-Allow-Origin", "*")

    config := getConfig()
    var buffer bytes.Buffer

    for _,Feature := range config.Features {
        var Toggle bool;
        if Feature.Persistent {
            cookie, err := req.Cookie("toogles-" + Feature.Name)

            if err == nil {
                Toggle = cookie.Value == "1"
            } else {
                Toggle = Feature.Toggle(req)

                Value := "0"
                if Toggle {
                    Value = "1"
                }
                cookie := http.Cookie{
                    Name: "toogles-" + Feature.Name,
                    Value: Value,
                    MaxAge: Feature.Expire,
                }
                http.SetCookie(res, &cookie)
            }
        } else {
            Toggle = Feature.Toggle(req)
        }

        if Toggle {
            buffer.WriteString(Feature.Name)
            buffer.WriteString(",")
        }
    }

    featuresString := strings.TrimRight(buffer.String(), ",")

    if len(featuresString) == 0 {
        fmt.Fprintf(res, "[]")
        return
    }

    features, err := json.Marshal(strings.Split(featuresString, ","))

    if err != nil {
        log.Fatal(err)
    }

    result := string(features[:])
    fmt.Fprintf(res, result)
}

func isAuthed(req *http.Request) bool {
    apiKey := os.Getenv("API_KEY")
    Query := req.URL.Query()

    if apiKey == "" {
        log.Print("API_KEY not set, not allowed to update features")
        return false
    }

    return apiKey == Query.Get("key")
}

func HandleFeatures(res http.ResponseWriter, req *http.Request) {
    defer req.Body.Close()

    if isAuthed(req) == false {
        res.WriteHeader(http.StatusUnauthorized)
        fmt.Fprint(res, "")

        return
    }
    config := getConfig()
    if req.Method == http.MethodGet {
        configBytes, _ := json.Marshal(config)
        configJson := string(configBytes[:])

        res.Header().Set("Content-Type", "application/json")
        res.Header().Set("Access-Control-Allow-Origin", "*")
        fmt.Fprintf(res, configJson)
    } else if req.Method == http.MethodPost {
        body, err := ioutil.ReadAll(req.Body)
        if err != nil {
            log.Fatal(err)
        }
        log.Print("Setting features")
        log.Print(string(body))
        setConfigString(string(body))

        saveConfigToRedis()
    }
}

func HandleFeature(res http.ResponseWriter, req *http.Request) {
    defer req.Body.Close()

    if isAuthed(req) == false {
        res.WriteHeader(http.StatusUnauthorized)
        fmt.Fprint(res, "")

        return
    }

    config := getConfig()
    if req.Method == http.MethodGet || req.Method == http.MethodPut {
        Query := req.URL.Query()
        name := Query.Get("name")

        if name == "" {
            res.WriteHeader(http.StatusNotFound)
            fmt.Fprint(res, "")

            return
        }

        var feature Feature
        var featureIndex int

        for i, _feature := range config.Features {
            if _feature.Name == name {
                feature = _feature
                featureIndex = i
                break
            }
        }

        if &feature == nil {
            res.WriteHeader(http.StatusNotFound)
            fmt.Fprint(res, "")

            return
        }

        if req.Method == http.MethodPut {
            body, err := ioutil.ReadAll(req.Body)
            if err != nil {
                log.Fatal(err)
            }

            if err := json.Unmarshal(body, &feature); err != nil {
                log.Fatal(err)
            }
        }

        config.Features[featureIndex] = feature

        saveConfigToRedis()

        res.Header().Set("Content-Type", "application/json")
        res.Header().Set("Access-Control-Allow-Origin", "*")

        featureBytes, _ := json.Marshal(feature)
        fmt.Fprintf(res, string(featureBytes[:]))
    } else if req.Method == http.MethodPost {
        body, err := ioutil.ReadAll(req.Body)
        if err != nil {
            log.Fatal(err)
        }
        var feature Feature
        if err := json.Unmarshal(body, &feature); err != nil {
            log.Fatal(err)
        }

        var _feature *Feature = nil
        for _, f := range config.Features {
            if f.Name == feature.Name {
                _feature = &f
                break
            }
        }

        if _feature != nil {
            res.WriteHeader(http.StatusConflict)
            fmt.Fprint(res, "Feature with the same name already exists: " + _feature.Name)

            return
        }

        config.Features = append(config.Features, feature)

        setConfig(config)
        saveConfigToRedis()

        res.Header().Set("Content-Type", "application/json")
        res.Header().Set("Access-Control-Allow-Origin", "*")

        featureBytes, _ := json.Marshal(feature)
        fmt.Fprintf(res, string(featureBytes[:]))
    }

    fmt.Fprint(res, "")
}

func HandleHealthCheck(res http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(res, "")
}

func HandleStats(res http.ResponseWriter, req *http.Request) {
    stats = GetStats()

    statsBytes, _ := json.Marshal(stats)
    res.Header().Set("Content-Type", "application/json")
    res.Header().Set("Access-Control-Allow-Origin", "*")
    fmt.Fprintf(res, string(statsBytes[:]))
}

func main() {
    loadConfigFromRedis()

    http.HandleFunc("/", HandleToggles)
    http.HandleFunc("/stats", HandleStats)
    http.HandleFunc("/health-check", HandleHealthCheck)
    http.HandleFunc("/features", HandleFeatures)
    http.HandleFunc("/feature", HandleFeature)

    server := http.Server{
        Addr: ":" + Port,
        ReadTimeout: 5 * time.Second,
        WriteTimeout: 5 * time.Second,
    }

    server.ListenAndServe()

    log.Printf("App is listening on port: %s", Port)
}
