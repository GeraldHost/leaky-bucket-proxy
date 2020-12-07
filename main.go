package main

import (
    "fmt"
    "net/http"
    "time"
    "errors"
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
  if _, ok := b.users[ip]; ok {
    user := b.users[ip]
    user.count++
    b.users[ip] = user
    if user.count >= 100 {
      return false
    }
  } else {
    b.users[ip] = User { ip: ip, count: INITIAL_COUNT }
  }
  return true
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
    allowed := bucket.Throttle(ip)

    if allowed {
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
  capacity := 100

  bucket := &Bucket { users, capacity }
  
  mux := http.NewServeMux()
  mux.Handle("/", handler(bucket))
  
  go scheduler(1 * time.Second, bucket)
  http.ListenAndServe(":8090", mux)
}


