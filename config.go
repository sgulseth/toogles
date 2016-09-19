package main

import (
    "fmt"
    "log"
    "time"
    "sync"
    "strings"
    "net/http"

    "encoding/json"
)


type Feature struct {
    Id              string           `json:"id"`
    Name            string           `json:"name"`
    Description     string           `json:"description"`
    Persistent      bool             `json:"persistent"`
    Expire          int              `json:"expire"`
    ShareStrategy   *ShareStrategy   `json:"shareStrategy,omitempty"`
    FirstStrategy   *FirstStrategy   `json:"firstStrategy,omitempty"`
    QueryStrategy   *QueryStrategy   `json:"queryStrategy,omitempty"`
    HeaderStrategy  *HeaderStrategy  `json:"headerStrategy,omitempty"`
    IPStrategy      *IPStrategy      `json:"ipStrategy,omitempty"`
    Stats           map[string]int64 `json:"-"`
}

func (self *Feature) Toggle(req *http.Request) bool {
    if self.ShareStrategy != nil {
        return self.ShareStrategy.Toggle(self, req)
    } else if self.FirstStrategy != nil {
        return self.FirstStrategy.Toggle(self, req)
    } else if self.QueryStrategy != nil {
        return self.QueryStrategy.Toggle(self, req)
    } else if self.HeaderStrategy != nil {
        return self.HeaderStrategy.Toggle(self, req)
    } else if self.IPStrategy != nil {
        return self.IPStrategy.Toggle(self, req)
    }

    return false
}

func (self *Feature) SetStat(stat string, val int64) {
    if self.Stats == nil {
        self.Stats = make(map[string]int64)
    }

    self.Stats[stat] = val
}

func (self *Feature) GetStat(stat string) int64 {
    if val, ok := self.Stats[stat]; ok {
        return val
    }

    return 0
}

var mutex *sync.Mutex
var features []Feature;

func getFeatures() []Feature {
    return features
}

func setFeatures(_features []Feature) []Feature {
    features = _features

    return features
}

func saveFeaturesToRedis() {
    featureBytes, _ := json.Marshal(features)
    featureJson := string(featureBytes[:])
    log.Print("Saving featureuration to redis")
    log.Print(featureJson)

    client := getRedisClient()
    err := client.Set("toogles_features", featureJson, 0).Err()

    if err != nil {
        log.Fatal(err)
    }
}

func loadFeaturesFromRedis() {
    client := getRedisClient()
    data, err := client.Get("toogles_features").Result()

    if err = json.Unmarshal([]byte(data), &features); err != nil {
        log.Print("Could not find features in redis")
        return
    }

    log.Print("Loaded features from redis")
    log.Print(data)
}

var FeatureStatsTypeMap = make(map[string]map[string]int64)

func IncrFeatureStatsType(featureId string, _type string) {
    mutex.Lock()
    if _, ok := FeatureStatsTypeMap[featureId]; ok == false {
        FeatureStatsTypeMap[featureId] = make(map[string]int64)
    }

    if _, ok := FeatureStatsTypeMap[featureId][_type]; ok == false {
        FeatureStatsTypeMap[featureId][_type] = 0
    }

    FeatureStatsTypeMap[featureId][_type]++
    mutex.Unlock()
}

func syncStatsRedis() {
    client := getRedisClient()
    t := time.Now()

    mutex.Lock()

    for id, statsMap := range FeatureStatsTypeMap {
        for statType, statVal := range statsMap {
            totalKey := fmt.Sprintf("toogles_features_%s_stats_%s_total", id, statType)
            client.IncrBy(totalKey, statVal)

            map_key := fmt.Sprintf("toogles_features_%s_stats_historic_%s", id, statType)

            minutesKey := fmt.Sprintf("%d_%d_%d_%d_%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
            client.HIncrBy(map_key + "_minutes", minutesKey, statVal)

            hoursKey := fmt.Sprintf("%d_%d_%d_%d", t.Year(), t.Month(), t.Day(), t.Hour())
            client.HIncrBy(map_key + "_hours", hoursKey, statVal)

            daysKey := fmt.Sprintf("%d_%d_%d", t.Year(), t.Month(), t.Day())
            client.HIncrBy(map_key + "_days", daysKey, statVal)
        }
    }

    FeatureStatsTypeMap = make(map[string]map[string]int64)

    for i, _ := range features {
        feature := &features[i]

        if feature.Stats == nil {
            feature.Stats = make(map[string]int64)
        }

        pattern := fmt.Sprintf("toogles_features_%s_stats_", feature.Id)
        keys, err := client.Keys(fmt.Sprintf("toogles_features_%s_stats_*", feature.Id)).Result()
        if err != nil {
            log.Print(err)
            continue
        }

        for _, key := range keys {
            value, err := client.Get(key).Int64()
            if err != nil {
                continue
            }

            //Users_total => [Users total]
            keyType := strings.Split(strings.Replace(key, pattern, "", -1), "_")

            feature.SetStat(keyType[0], value)
        }
    }


    mutex.Unlock()
}

type FeatureStats struct {
    FeatureId   string           `json:"featureId,omitempty"`
    Stats       map[string]int64 `json:"stats"`
}

func getFeatureStats(feature Feature) FeatureStats {
    return FeatureStats {
        FeatureId: feature.Id,
        Stats: feature.Stats,
    }
}

func deleteFeatureStatsFromRedis(featureId string) {
    client := getRedisClient()
    keys, err := client.Keys(fmt.Sprintf("toogles_features_%s_stats_", featureId) + "*").Result()

    if err != nil {
        return
    }

    for _, key := range keys {
        err := client.Del(key).Err()
        if err != nil {
            continue
        }
    }
}
