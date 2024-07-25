package main

import (
	"fmt"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/transport"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
)

func main() {
	bettyTransport, err := transport.NewQuicTransport("localhost:4242")
	fmt.Println("Will this run?")
	fmt.Println(*bettyTransport)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Betty has been set up!")

	for bettyTransport.AcceptedStreams[7] == nil {
		time.Sleep(time.Second * 1)
	}

	err = bettyTransport.Send([]byte{255, 254, 253, 252}, wgpstypes.DataChannel, wgpstypes.SyncRoleBetty)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Betty has sent a message!")
	err = bettyTransport.Send([]byte{251, 250}, wgpstypes.DataChannel, wgpstypes.SyncRoleBetty)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Betty has sent a message!")
	err = bettyTransport.Send([]byte{249}, wgpstypes.DataChannel, wgpstypes.SyncRoleBetty)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Betty has sent a message!")
	time.Sleep(time.Second * 3)

	bettyMessage, err := bettyTransport.Recv(wgpstypes.DataChannel, wgpstypes.SyncRoleBetty)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(bettyMessage)
	return

}
