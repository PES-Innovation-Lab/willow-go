package main

import (
	"fmt"

	pinagoladastore "github.com/PES-Innovation-Lab/willow-go/PinaGoladaStore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
)

func main() {
	/* bettyTransport, err := transport.NewQuicTransport("localhost:4242")
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
	return */

	WillowStore := (*pinagoladastore.InitStorage(types.NamespaceId("myspace")))
	newMessengerChan := make(chan wgps.NewMessengerReturn[string, types.SubspaceId, string, string, string, int, string, types.SubspaceId, string, string, string, string, string, string, string, []byte, uint], 1)
	opts := wgps.WgpsMessengerOpts[string, types.SubspaceId, string, string, string, int, string, types.SubspaceId, string, string, string, string, string, string, string, []byte, uint]{
		Schemes: wgpstypes.SyncSchemes[
			string,
			types.SubspaceId,
			string,
			string,
			string,
			int,
			string,
			types.SubspaceId,
			string,
			string,
			string,
			string,
			string,
			string,
			string,
			[]byte,
			uint,
		]{
			NamespaceScheme: pinagoladastore.TestNameSpaceScheme,
			SubspaceScheme:  pinagoladastore.TestSubspaceScheme,
			Payload:         pinagoladastore.TestPayloadScheme,
			Fingerprint:     pinagoladastore.TestFingerprintScheme,
			PathParams:      pinagoladastore.TestPathParams,
		},
		GetStore: func(namespace types.NamespaceId) store.Store[string, string, uint, []byte, string] {
			return WillowStore
		},
	}

	go wgps.NewWgpsMessenger(opts, newMessengerChan, "localhost:4241")
	messenger := <-newMessengerChan
	if messenger.Error != nil {
		fmt.Println("Error in creating messenger:", messenger.Error)
		return
	}
	fmt.Println("Messenger set up")
	messenger.NewMessenger.Initiate("localhost:4242")
}
