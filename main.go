package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type MyChannel struct {
	C      chan int
	closed bool
	mutex  sync.Mutex
}

func NewMyChannel() *MyChannel {
	return &MyChannel{C: make(chan int)}
}

func (mc *MyChannel) SafeClose() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	if !mc.closed {
		close(mc.C)
		mc.closed = true
	}
}

func (mc *MyChannel) IsClosed() bool {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	return mc.closed
}

/*
	Сделать функцию merge которая принимает список каналов int и возвращает результирующий канал
*/
func main() {
	ch1 := NewMyChannel()
	ch2 := NewMyChannel()
	ch3 := NewMyChannel()
	ch4 := NewMyChannel()

	ctx, cancel := context.WithCancel(context.Background())

	mergedChan := mergeChannels(ctx, ch1, ch2, ch3, ch4)

	go func() {
		ch1.C<- 1
		ch3.C<- 10
		ch2.C<- 15
		ch4.C<- 100
		ch2.C<- 134

		ch1.SafeClose()
		ch2.SafeClose()
		ch3.SafeClose()
		ch4.SafeClose()
	}()

	go func() {
		time.Sleep(time.Second * 3)
		cancel()
	}()

	for msg := range mergedChan {
		log.Println(msg)
	}
}

func mergeChannels(ctx context.Context, chs ...*MyChannel) chan int {
	resultCh := make(chan int)

	wg := sync.WaitGroup{}
	for _, ch := range chs {
		wg.Add(1)
		go func(ch *MyChannel) {
			for {
				select {
				case <-ctx.Done():
					{
						ch.SafeClose()
						break
					}
				case msg, ok := <-ch.C:
					{
						if ok {
							resultCh<- msg
						} else {
							if ch.IsClosed() {
								break
							}
						}
					}
				default:
					continue
				}

				if ch.IsClosed() {
					wg.Done()
					break
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh
}
