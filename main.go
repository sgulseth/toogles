package main

import (
    "os"
    "fmt"
    "time"
    "sync"
    "strings"
    "bytes"
    "net/http"
    "log"
    "io/ioutil"
    "encoding/json"
)

var PortEnv = GetEnv("PORT", "8080")
var CorsEnv = GetEnv("CORS", "*")

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
    SetCorsHeaders(res, req, CorsEnv, "GET")

    features := getFeatures()
    var buffer bytes.Buffer

    for _,Feature := range features {
        var Toggle bool;
        if Feature.Persistent {
            cookie, err := req.Cookie("toogles-" + Feature.Id)

            if err == nil {
                Toggle = cookie.Value == "1"
            } else {
                Toggle = Feature.Toggle(req)

                Value := "0"
                if Toggle {
                    Value = "1"
                }
                cookie := http.Cookie{
                    Name: "toogles-" + Feature.Id,
                    Value: Value,
                    MaxAge: Feature.Expire,
                }
                http.SetCookie(res, &cookie)

            }
        } else {
            Toggle = Feature.Toggle(req)
        }

        if Toggle {
            buffer.WriteString(Feature.Id)
            buffer.WriteString(",")
            IncrFeatureStatsType(Feature.Id, "impressions")
        }
    }

    featuresString := strings.TrimRight(buffer.String(), ",")

    if len(featuresString) == 0 {
        fmt.Fprintf(res, "[]")
        return
    }

    response, err := json.Marshal(strings.Split(featuresString, ","))

    if err != nil {
        log.Fatal(err)
    }

    fmt.Fprintf(res, string(response[:]))
}

func isAuthed(req *http.Request) bool {
    apiKey := os.Getenv("API_KEY")
    BearerToken := "Bearer " + apiKey
    if len(req.Header["Authorization"]) == 0 {
        return false
    }

    Header := req.Header["Authorization"][0]

    if apiKey == "" {
        log.Print("API_KEY not set, not allowed to update features")
        return false
    }

    return BearerToken == Header
}

func HandleHealthCheck(res http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(res, "")
}

func HandleStats(res http.ResponseWriter, req *http.Request) {
    stats = GetStats()

    statsBytes, _ := json.Marshal(stats)
    res.Header().Set("Content-Type", "application/json")
    SetCorsHeaders(res, req, CorsEnv, "GET")

    fmt.Fprintf(res, string(statsBytes[:]))
}

func HandleApiFeatures(res http.ResponseWriter, req *http.Request) {
    defer req.Body.Close()

    res.Header().Set("Content-Type", "application/json")
    SetCorsHeaders(res, req, CorsEnv, "GET")

    if req.Method == http.MethodOptions {
        fmt.Fprint(res, "")
        return
    }

    if isAuthed(req) == false {
        res.WriteHeader(http.StatusUnauthorized)
        fmt.Fprint(res, "")

        return
    }

    features := getFeatures()
    featureBytes, _ := json.Marshal(features)
    fmt.Fprintf(res, string(featureBytes[:]))
}

func HandleApiFeaturesStats(res http.ResponseWriter, req *http.Request) {
    defer req.Body.Close()

    res.Header().Set("Content-Type", "application/json")
    SetCorsHeaders(res, req, CorsEnv, "GET")

    if req.Method == http.MethodOptions {
        fmt.Fprint(res, "")

        return
    }

    if isAuthed(req) == false {
        res.WriteHeader(http.StatusUnauthorized)
        fmt.Fprint(res, "")

        return
    }

    features := getFeatures()
    stats := make([]FeatureStats, len(features))

    for i, feature := range features {
        stats[i] = getFeatureStats(feature)
    }

    json, _ := json.Marshal(stats)
    fmt.Fprintf(res, string(json[:]))
}

func HandleApiFeature(res http.ResponseWriter, req *http.Request) {
    defer req.Body.Close()

    res.Header().Set("Content-Type", "application/json")
    SetCorsHeaders(res, req, CorsEnv, "POST,PUT,GET,DELETE")

    if req.Method == http.MethodOptions {
        fmt.Fprint(res, "")

        return
    }

    if isAuthed(req) == false {
        res.WriteHeader(http.StatusUnauthorized)
        fmt.Fprint(res, "")

        return
    }

    features := getFeatures()
    if req.Method == http.MethodGet || req.Method == http.MethodPut || req.Method == http.MethodDelete {
        Query := req.URL.Query()
        id := Query.Get("id")

        if id == "" {
            res.WriteHeader(http.StatusNotFound)
            fmt.Fprint(res, "")

            return
        }

        var feature Feature
        var featureIndex int

        for i, _feature := range features {
            if _feature.Id == id {
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

        if req.Method == http.MethodDelete {
            fmt.Fprint(res, "{}")

            features = append(features[:featureIndex], features[featureIndex+1:]...)
            setFeatures(features)
            saveFeaturesToRedis()
            deleteFeatureStatsFromRedis(feature.Id)
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

        features[featureIndex] = feature

        saveFeaturesToRedis()

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

        id := slug(feature.Name)

        var _feature *Feature = nil
        for _, f := range features {
            if f.Id == id {
                _feature = &f
                break
            }
        }

        if _feature != nil {
            res.WriteHeader(http.StatusConflict)
            fmt.Fprint(res, "Feature with the same name/id already exists: " + _feature.Name)

            return
        }

        feature.Id = id

        features = append(features, feature)

        setFeatures(features)
        saveFeaturesToRedis()

        featureBytes, _ := json.Marshal(feature)
        fmt.Fprintf(res, string(featureBytes[:]))
    } else {
        fmt.Fprint(res, "")
    }
}

func startServer() {
    http.HandleFunc("/", HandleToggles)
    http.HandleFunc("/stats", HandleStats)
    http.HandleFunc("/health-check", HandleHealthCheck)
    http.HandleFunc("/api/features", HandleApiFeatures)
    http.HandleFunc("/api/features/stats", HandleApiFeaturesStats)
    http.HandleFunc("/api/feature", HandleApiFeature)

    server := http.Server{
        Addr: ":" + PortEnv,
        ReadTimeout: 5 * time.Second,
        WriteTimeout: 5 * time.Second,
    }

    log.Printf("App is listening on port: %s", PortEnv)

    server.ListenAndServe()
}

func startStatsSync() {
    t := time.NewTicker(time.Second * 5)

    for {
        syncStatsRedis()
        <-t.C
    }
}

func main() {
    mutex = &sync.Mutex{}

    loadFeaturesFromRedis()
    syncStatsRedis()

    go startStatsSync()

    startServer()
}
