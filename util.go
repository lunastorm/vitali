package vitali

import (
    "fmt"
    "strings"
    "runtime"
    "hash/crc32"
)

func panicOnErr(res interface{}, e error) interface{} {
    if e != nil {
        panic(e)
    }   
    return res 
}

func errorCode(msg string) uint32 {
    return crc32.ChecksumIEEE([]byte(msg))
}

func lineInfo(skip int) (where string) {
    _, fn, ln, ok := runtime.Caller(skip)
    if ok {
        where = fmt.Sprintf("%s:%d", fn, ln)
    }
    return
}

func fullTrace(skip int, sep string) (trace string) {
    for {
        where := lineInfo(skip)
        skip++
        if where == "" {
            return
        } else if strings.HasSuffix(where, ":0") {
            continue
        }
        trace = trace + sep + where
    }
    return
}
