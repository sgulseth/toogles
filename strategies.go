package main

import (
    "net/http"
    "math/rand"
)

type Strategy interface {
    Toggle(req *http.Request)    bool
}

type ShareStrategy struct {
    Share       int
}

func (self *ShareStrategy) Toggle(req *http.Request) bool {
    r := rand.Intn(100)
    return self.Share > r
}

type FirstStrategy struct {
    First       int
    Current     int
}

func (self *FirstStrategy) Toggle(req *http.Request) bool {
    if self.Current < self.First {
        self.Current++

        return true
    }

    return false
}

type QueryStrategy struct {
    Key         string
    Value       string
    Values      []string
}

func (self *QueryStrategy) Toggle(req *http.Request) bool {
    URL := req.URL
    Query := URL.Query()
    Value := Query.Get(self.Key)
    if Value != "" {
        if self.Values != nil {
            return contains(self.Values, Value)
        } else {
            return Value == self.Value
        }
    }

    return false
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}