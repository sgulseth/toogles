package main

import (
    "net"
    "net/http"
    "math/rand"
)

type Strategy interface {
    Toggle(feature *Feature, req *http.Request)    bool
}

type ShareStrategy struct {
    Share       int         `json:"share"`
}

func (self *ShareStrategy) Toggle(feature *Feature, req *http.Request) bool {
    r := rand.Intn(100)
    return self.Share > r
}

type FirstStrategy struct {
    First       int64         `json:"first"`
}

func (self *FirstStrategy) Toggle(feature *Feature, req *http.Request) bool {
    if feature.GetStat("Users") < self.First {
        return true
    }

    return false
}

type QueryStrategy struct {
    Key         string     `json:"key"`
    Values      []string   `json:"values"`
}

func (self *QueryStrategy) Toggle(feature *Feature, req *http.Request) bool {
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

func (self *HeaderStrategy) Toggle(feature *Feature, req *http.Request) bool {
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

func (self *IPStrategy) Toggle(feature *Feature, req *http.Request) bool {
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