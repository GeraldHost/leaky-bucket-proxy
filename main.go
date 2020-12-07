package main

import (
    "fmt"
    "net/http"
    "time"
    "errors"
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

func handler(bucket *Bucket) http.Handler { 
  return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    ip := req.RemoteAddr
    throttled := bucket.Throttle(ip)
    if !throttled {
      fmt.Fprint(w, "Welcome to the website")
    } else {
      fmt.Fprint(w, "You have made too many requests")
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

func parseArgs() (int, int, error) {
  if len(os.Args) < 3 {
    return 0, 0, errors.New("Not enough args")
  }
  capacity, e1 := strconv.Atoi(os.Args[1])
  refresh, e2 := strconv.Atoi(os.Args[2])
  if e1 != nil || e2 != nil {
    return 0, 0, errors.New("Please provide numbers")
  }
  return capacity, refresh, nil
}

func main() {
  capacity, refreshSeconds, err := parseArgs()
  if err != nil {
    fmt.Println(err)
    return
  }
  users := make(map[string]User)
  bucket := &Bucket { users, capacity }
  mux := http.NewServeMux()
  mux.Handle("/", handler(bucket))
  go scheduler(time.Duration(refreshSeconds) * time.Second, bucket)
  http.ListenAndServe(":8090", mux)
}


