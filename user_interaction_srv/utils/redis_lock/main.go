package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
)

func main() {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: "localhost:6379",
	})
	pool := goredis.NewPool(client)

	rs := redsync.New(pool)

	mutexname := "421"
	mutex := rs.NewMutex(mutexname)

	goNum := 2
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < goNum; i++ {
		go func() {
			defer wg.Done()
			fmt.Println("准备获取锁")
			if err := mutex.Lock(); err != nil {
				panic(err)
			}
			fmt.Println("获取锁成功")

			time.Sleep(time.Second * 5)

			fmt.Println("准备释放锁")
			if ok, err := mutex.Unlock(); !ok || err != nil {
				panic("unlock failed")
			}
			fmt.Println("释放锁成功")
		}()
	}
	wg.Wait()
}
