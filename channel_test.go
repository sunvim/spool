package spool

import (
	"net"
	"testing"
	"time"
)

func Test_NewChannelPool(t *testing.T) {
	servAddr := "127.0.0.1:7530"
	fac := func() (net.Conn, error) { return net.Dial("tcp", servAddr) }
	pool, err := NewChannelPool(2, 20, fac)
	if err != nil {
		t.Fatal("happen falied:", err)
	}
	t.Log("init pool num:", pool.Len())
	conOne, err := pool.Get()
	if err != nil {
		t.Fatal("happen falied:", err)
	}
	t.Log("current pool num:", pool.Len())
	conTwo, err := pool.Get()
	if err != nil {
		t.Fatal("happen falied:", err)
	}
	t.Log("current pool num:", pool.Len())
	conThree, err := pool.Get()
	if err != nil {
		t.Fatal("happen falied:", err)
	}
	t.Log("current pool num:", pool.Len())
	conFour, err := pool.Get()
	if err != nil {
		t.Fatal("happen falied:", err)
	}
	t.Log("current pool num:", pool.Len())
	time.Sleep(3 * time.Second)
	t.Log(" 1 time current pool num:", pool.Len())

	t.Log("close connect: put them back to pool")
	conOne.Close()
	conTwo.Close()
	conThree.Close()
	conFour.Close()
	time.Sleep(3 * time.Second)
	t.Log("2 time current pool num:", pool.Len())
	time.Sleep(3 * time.Second)
	t.Log("3 time current pool num:", pool.Len())
	time.Sleep(3 * time.Second)
	t.Log("4 time current pool num:", pool.Len())
	time.Sleep(3 * time.Second)
	t.Log("5 time current pool num:", pool.Len())
}
