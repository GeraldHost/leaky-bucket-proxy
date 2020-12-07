package main

import (
    "fmt"
    "net/http"
    "time"
    "errors"
)

const INITIAL_COUNT int = 0
const DEFAULT_CAPACITY int = 10
const DEFAULT_REFRESH = time.Second * 2

type User struct {
  ip string
  count int
}

type Bucket struct {
  users map[string]User
  capacity int
}

func (b Bucket) Throttle(ip string) bool {
  if _, ok := b.users[ip]; ok {
    user := b.users[ip]
    if v := user.count+1; v > b.capacity {
      user.count = b.capacity
    } else {
      user.count++
    }
    b.users[ip] = user
  } else {
    b.users[ip] = User { ip: ip, count: INITIAL_COUNT }
  }
  return b.users[ip].count >= b.capacity 
}

func (b Bucket) User(ip string) (User, error) {
  if user, ok := b.users[ip]; ok {
    return user, nil
  }
  return User{}, errors.New("User not found")
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

func main() {
  users := make(map[string]User)
  capacity := DEFAULT_CAPACITY

  bucket := &Bucket { users, capacity }
  
  mux := http.NewServeMux()
  mux.Handle("/", handler(bucket))
  
  go scheduler(DEFAULT_REFRESH, bucket)
  http.ListenAndServe(":8090", mux)
}


