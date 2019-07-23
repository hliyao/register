package lock

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMutex1(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	InitEtcdv3Client()
	time.Sleep(2 * time.Second)
	var eLock Etcdv3Lock
	fmt.Println("begin")
	err := eLock.Init("/mylock")
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return
	}

	err = eLock.Lock(ctx)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return
	}
	fmt.Println("acquired lock for s1")

	time.Sleep(500 * time.Second)

	/*err = eLock.UnLock(ctx)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return
	}*/
}

func TestMutex2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	InitEtcdv3Client()
	time.Sleep(2 * time.Second)
	var eLock Etcdv3Lock
	err := eLock.Init("/mylock")
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return
	}
	err = eLock.Lock(ctx)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return
	}
	fmt.Println("acquired lock for s1")

	err = eLock.UnLock(ctx)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return
	}
	fmt.Println("unlock for s1")
	time.Sleep(500 * time.Second)
}
