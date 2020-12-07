package main

import (
    "fmt"
    "net/http"
    "net/http/httputil"
    "net/url"
    "time"
    "os"
    "strconv"
)

const INITIAL_COUNT int = 0

type User struct {
  ip string
  count int
}

type Bucket struct {
  users map[string]User
  capacity int
}

func (b Bucket) Throttle(ip string) bool {
  if user, ok := b.users[ip]; ok {
    if !(user.count >= b.capacity) {
      user.count++
    }
    b.users[ip] = user
  } else {
    b.users[ip] = User { ip: ip, count: INITIAL_COUNT }
  }
  return b.users[ip].count >= b.capacity 
}

func (b Bucket) Fill() {
  for i, user := range b.users {
    if v := user.count - 1; v < INITIAL_COUNT {
      user.count = INITIAL_COUNT
    } else {
      user.count--
    }
    b.users[i] = user
  }
}

func handler(bucket *Bucket, proxy *httputil.ReverseProxy) http.Handler { 
  return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
    ip := req.RemoteAddr
    throttled := bucket.Throttle(ip)
    if !throttled {
      proxy.ServeHTTP(res, req)
    } else {
      fmt.Fprint(res, "You have made too many requests")
    }
  })
}

func scheduler(tick time.Duration, bucket *Bucket) {
  ticker := time.NewTicker(tick)
  defer ticker.Stop()
  for {
    select {
      case <-ticker.C:
        bucket.Fill()
        fmt.Println(bucket.users)
    }
  }
}

func parseArgs() (int, int, *url.URL) {
  capacity, _ := strconv.Atoi(os.Args[1])
  refresh, _ := strconv.Atoi(os.Args[2])

  forwardUrl, _ := url.Parse(os.Args[3])
  return capacity, refresh, forwardUrl
}

func main() {
  capacity, refreshSeconds, forwardUrl := parseArgs()
  proxy := httputil.NewSingleHostReverseProxy(forwardUrl)
  users := make(map[string]User)
  bucket := &Bucket { users, capacity }
  mux := http.NewServeMux()
  mux.Handle("/", handler(bucket, proxy))
  go scheduler(time.Duration(refreshSeconds) * time.Second, bucket)
  http.ListenAndServe(":8090", mux)
}


