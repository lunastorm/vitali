package util

import (
    "fmt"
    "sync"
)

type ChanMap struct {
    m map[string]chan int
    lock sync.Mutex
}

func CreateChanMap() (*ChanMap) {
    return &ChanMap{
        make(map[string]chan int),
        sync.Mutex{},
    }
}

func (c *ChanMap) Get(username string, slide string) (chan int) {
    c.lock.Lock()
    defer c.lock.Unlock()

    key := fmt.Sprintf("%s:%s", username, slide)
    if c.m[key] == nil {
        res := make(chan int)
        c.m[key] = res
        return res
    } else {
        return c.m[key]
    }
}
