package natspbufclient

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/nats-io/nats"
	pb "github.com/wvell/natspbufclient/testdata"
)

func NewConn() *nats.Conn {
	for {
		conn, err := nats.Connect("nats://127.0.0.1:4222")
		if err == nil {
			return conn
		}

		time.Sleep(time.Millisecond * 50)
	}

	return nil
}

func TestMain(m *testing.M) {
	cmd := exec.Command("gnatsd")
	go cmd.Run()

	m.Run()

	cmd.Process.Kill()
}

func NewProtoEncodedConn(t *testing.T) *nats.EncodedConn {
	ec, err := nats.NewEncodedConn(NewConn(), ENC_NAME)
	if err != nil {
		t.Fatalf("Failed to create an encoded connection: %v\n", err)
	}
	return ec
}

func TestProtoMarshalStruct(t *testing.T) {
	ec := NewProtoEncodedConn(t)
	defer ec.Close()
	ch := make(chan bool)

	me := &pb.Person{Name: "derek", Age: 22, Address: "85 Second St"}
	me.Children = make(map[string]*pb.Person)

	me.Children["sam"] = &pb.Person{Name: "sam", Age: 16, Address: "85 Second St"}
	me.Children["meg"] = &pb.Person{Name: "meg", Age: 14, Address: "85 Second St"}

	ec.Subscribe("proto_struct", func(p *pb.Person) {
		ch <- true
		if !reflect.DeepEqual(p, me) {
			t.Fatalf("Did not receive the correct struct response")
		}

		ch <- true
	})

	ec.Publish("proto_struct", me)
	if e := wait(ch); e != nil {
		t.Fatal("Did not receive the message")
	}
}

func wait(ch chan bool) error {
	return waitTime(ch, 500*time.Millisecond)
}

func waitTime(ch chan bool, timeout time.Duration) error {
	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
	}
	return errors.New("timeout")
}
