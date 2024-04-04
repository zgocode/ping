package ping

import (
	"fmt"
	"testing"
)

func TestRequestTest(t *testing.T) {
	ping := New("229.6.6.61", 0, 2500)
	fmt.Println(ping.Get())
	ping.conn.Close()
	fmt.Println(ping.Get())
	fmt.Println(ping.Get())
}
