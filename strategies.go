package main

import (
    "net"
    "time"
    "strings"
    "strconv"
    "net/http"
    "math/rand"
)

type Strategy interface {
    Toggle(feature *Feature, res http.ResponseWriter, req *http.Request)    bool
}

type ShareStrategy struct {
    Share       int         `json:"share"`
}

func (self *ShareStrategy) Toggle(feature *Feature, res http.ResponseWriter, req *http.Request) bool {
    if self.Share == 100 {
        return true
    }

    shareCookie, err := req.Cookie("toogles-" + feature.Id + "|share")
    var share int

    if err != nil {
        share := rand.Intn(100)
        shareString := strconv.Itoa(int(share))

        cookie := http.Cookie{
            Name: "toogles-" + feature.Id + "|share",
            Value: shareString,
        }
        http.SetCookie(res, &cookie)
    } else {
        s, _ := strconv.ParseInt(shareCookie.Value, 10, 64)
        share = int(s)
    }


    return self.Share <= share
}

type FirstStrategy struct {
    First       int64         `json:"first"`
}

func (self *FirstStrategy) Toggle(feature *Feature, res http.ResponseWriter, req *http.Request) bool {
    if feature.GetStat("impressions") < self.First {
        return true
    }

    return false
}

type QueryStrategy struct {
    Key         string     `json:"key"`
    Values      []string   `json:"values"`
}

func (self *QueryStrategy) Toggle(feature *Feature, res http.ResponseWriter, req *http.Request) bool {
    URL := req.URL
    Query := URL.Query()
    Value := Query.Get(self.Key)
    if Value != "" {
        return sliceContains(self.Values, Value)
    }

    return false
}

type HeaderStrategy struct {
    Header      string     `json:"header"`
    Values      []string   `json:"values"`
}

func (self *HeaderStrategy) Toggle(feature *Feature, res http.ResponseWriter, req *http.Request) bool {
    if _, ok := req.Header[self.Header]; ok == false {
        return false
    }

    if len(req.Header[self.Header]) > 0 {
        Value := req.Header[self.Header][0]
        return sliceContains(self.Values, Value)
    }

    return false
}

type IPStrategy struct {
    IPs      []string   `json:"ips"`
}

func (self *IPStrategy) Toggle(feature *Feature, res http.ResponseWriter, req *http.Request) bool {
    ip, _, err := net.SplitHostPort(req.RemoteAddr)

    if err != nil {
        return false
    }

    if (sliceContains(self.IPs, ip)) {
        return true
    }

    if forwardedForHeaders, ok := req.Header["X-Forwarded-For"]; ok  {
        for _, ip := range forwardedForHeaders {
            if sliceContains(self.IPs, ip) {
                return true
            }
        }
    }

    return false
}

type RefererStrategy struct {
    Referers      []string   `json:"referers"`
}

func (self *RefererStrategy) Toggle(feature *Feature, res http.ResponseWriter, req *http.Request) bool {
    if referers, ok := req.Header["Referer"]; ok {
        for _, referer := range referers {
            if sliceContains(self.Referers, referer) {
                return true
            }
        }
    }

    return false
}

type RetargetStrategy struct {
    FeatureIds      []string   `json:"featureIds"`
}

func (self *RetargetStrategy) Toggle(feature *Feature, res http.ResponseWriter, req *http.Request) bool {
    for _, featureId := range self.FeatureIds {
        if _, err := req.Cookie("toogles-" + featureId); err != nil {
            return true
        }
    }

    return false
}

type UserRecurrencyStrategy struct {
    Visits             int   `json:"visits"`
    VisitInterval      int   `json:"visitInterval"`
}

func (self *UserRecurrencyStrategy) Toggle(feature *Feature, res http.ResponseWriter, req *http.Request) bool {
    cookie, err := req.Cookie("toogles-" + feature.Id + "|visits")
    visits := 1
    now := time.Now().Unix()

    if err == nil {
        value := strings.Split(cookie.Value, "|")
        if len(value) != 2 {
            return false
        }

        visits64, _ := strconv.ParseInt(value[0], 10, 64)
        visits = int(visits64)
        lastVisit, _ := strconv.ParseInt(value[1], 10, 64)

        diff := int(now - lastVisit)
        if diff > self.VisitInterval {
            visits += 1
        }

        if visits >= self.Visits {
            return true
        }
    }

    visitsString := strconv.Itoa(visits)
    nowString := strconv.FormatInt(now, 10)


    visitCookie := http.Cookie{
        Name: "toogles-" + feature.Id + "|visits",
        Value: visitsString + "|" + nowString,
    }

    http.SetCookie(res, &visitCookie)

    return false
}