package main

import (
    "os"
    "fmt"
    "strings"
    "bytes"
    "net/http"
    "log"
    "io/ioutil"
    "encoding/json"
)

func HandleToggles(res http.ResponseWriter, req *http.Request) {
    if req.URL.Path != "/" {
        res.WriteHeader(http.StatusNotFound)
        return
    }

    config := getConfig()
    var buffer bytes.Buffer

    for _,Feature := range config.Features {
        var Toggle bool;
        if Feature.Persistent {
            cookie, err := req.Cookie(Feature.Name)

            if err == nil {
                Toggle = cookie.Value == "1"
            } else {
                Toggle = Feature.Toggle(req)

                Value := "0"
                if Toggle {
                    Value = "1"
                }
                cookie := http.Cookie{
                    Name: Feature.Name,
                    Value: Value,
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

    res.Header().Set("Content-Type", "application/json")
    res.Header().Set("Access-Control-Allow-Origin", "*")

    result := string(features[:])
    fmt.Fprintf(res, result)

    StatsIncrementConnections()
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
    /*shareStrategy := ShareStrategy{
        Share: 50,
    }
    dummyFeature1 := Feature{
        Name: "half-n-half",
        Persistent: true,
        ShareStrategy: &shareStrategy,
    }

    firstStrategy := FirstStrategy{
        First: 3,
    }
    dummyFeature2 := Feature{
        Name: "first-users",
        Persistent: true,
        FirstStrategy: &firstStrategy,
    }

    userStrategy := QueryStrategy{
        Key: "user-id",
        Values: []string{"1234", "5678"},
    }
    dummyFeature3 := Feature{
        Name: "users-ids",
        Persistent: false,
        QueryStrategy: &userStrategy,
    }

    config := Configuration{
        Features: []Feature {
            dummyFeature1,
            dummyFeature2,
            dummyFeature3,
        },
    }

    setConfig(config)*/

    loadConfigFromRedis()

    http.HandleFunc("/", HandleToggles)
    http.HandleFunc("/stats", HandleStats)
    http.HandleFunc("/health-check", HandleHealthCheck)
    http.HandleFunc("/features", HandleFeatures)
    log.Fatal(http.ListenAndServe(":8080", nil))
}