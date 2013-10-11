package vitali

import (
    "fmt"
    "strings"
    "runtime"
    "hash/crc32"
)

func errorCode(msg string) uint32 {
    return crc32.ChecksumIEEE([]byte(msg))
}

func lineInfo(skip int) (where string) {
    _, fn, ln, ok := runtime.Caller(skip)
    if ok {
        where = fmt.Sprintf("%s:%d", strings.SplitN(fn, "/src/", 2)[1], ln)
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
