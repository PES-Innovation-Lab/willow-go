package tests

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/transport"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
)

func TestQuicTransportSendAndReceive(t *testing.T) {
	var alfieTransport *transport.QuicTransport
	//var bettyTransport *transport.QuicTransport
	var err error
	// Create a new QuicTransport
	/*go func() {
		bettyTransport, err = transport.NewQuicTransport(wgpstypes.SyncRoleBetty, "localhost:4242")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Betty has been set up!")
	}() */
	//time.Sleep(time.Second * 10)
	// Create a new QuicTransport
	alfieTransport, err = transport.NewQuicTransport("localhost:4241")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Alfie has been set up!")
	err = alfieTransport.Initiate("localhost:4242")
	if err != nil {
		t.Fatal(err)
	}

	// Send a message from Alfie to Betty
	err = alfieTransport.Send([]byte{1, 2, 3, 4}, wgpstypes.DataChannel, wgpstypes.SyncRoleAlfie)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Alfie has sent a message!")

	err = alfieTransport.Send([]byte{5, 6}, wgpstypes.DataChannel, wgpstypes.SyncRoleAlfie)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Alfie has sent a message!")

	err = alfieTransport.Send([]byte{7}, wgpstypes.DataChannel, wgpstypes.SyncRoleAlfie)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Alfie has sent a message!")

	var alfieMessage []byte

	// Receive a message from Alfie to Betty
	time.Sleep(time.Second * 10)
	alfieMessage, err = alfieTransport.Recv(wgpstypes.DataChannel, wgpstypes.SyncRoleAlfie)
	if err != nil {
		t.Fatal(err)
	}

	//fmt.Println(string(bettyMessage))
	fmt.Println(alfieMessage)
	if !reflect.DeepEqual(alfieMessage, []byte{255, 254, 253, 252, 251, 250, 249}) {
		t.Errorf("Did not receive correct message. Wanted %v, got %v", []byte{255, 254, 253, 252, 251, 250, 249}, alfieMessage)
	}

}
