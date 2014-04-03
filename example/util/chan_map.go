package util

import (
    "fmt"
    "sync"
)

type ChanMap struct {
    m map[string]map[string]chan int
    lock sync.Mutex
}

func CreateChanMap() (*ChanMap) {
    return &ChanMap{
        make(map[string]map[string]chan int),
        sync.Mutex{},
    }
}

func (c *ChanMap) Get(username string, slide string, remoteAddr string) (chan int) {
    c.lock.Lock()
    defer c.lock.Unlock()

    key := fmt.Sprintf("%s:%s", username, slide)
    if c.m[key] == nil {
        c.m[key] = make(map[string]chan int)
    }
    c.m[key][remoteAddr] = make(chan int)
    return c.m[key][remoteAddr]
}

func (c *ChanMap) Remove(username string, slide string, remoteAddr string) {
    c.lock.Lock()
    defer c.lock.Unlock()

    key := fmt.Sprintf("%s:%s", username, slide)
    delete(c.m[key], remoteAddr)
    if len(c.m[key]) == 0 {
        delete(c.m, key)
    }
}

func (c *ChanMap) Broadcast(username string, slide string, page int) {
    c.lock.Lock()
    defer c.lock.Unlock()

    key := fmt.Sprintf("%s:%s", username, slide)
    for _, ch := range c.m[key] {
        ch <- page
    }
}
