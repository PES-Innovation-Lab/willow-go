package main

import (
	"fmt"
	"time"

	pinagoladastore "github.com/PES-Innovation-Lab/willow-go/PinaGoladaStore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
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
	hello1 := "Hello, world!"
	var hello1Bytes []byte
	hello1Bytes = append(hello1Bytes, utils.BigIntToBytes(uint64(len(hello1)))...)
	hello1Bytes = append(hello1Bytes, []byte(hello1)...)
	fmt.Printf("Sending %v now!!!\n", hello1Bytes)
	//r := bytes.NewReader(hello1Bytes)
	//// := binary.BigEndian.Uint64(hello1Bytes[:8])

	//fmt.Printf("Length of hello1Bytes is %v\n", intVal)

	hello2 := "Hello, world 2!"
	hello3 := "Hello, world 3!"
	var hello2Bytes []byte
	hello2Bytes = append(hello2Bytes, utils.BigIntToBytes(uint64(len(hello2)))...)
	hello2Bytes = append(hello2Bytes, []byte(hello2)...)
	var hello3Bytes []byte
	hello3Bytes = append(hello3Bytes, utils.BigIntToBytes(uint64(len(hello3)))...)
	hello3Bytes = append(hello3Bytes, []byte(hello3)...)
	messenger.NewMessenger.Transport.Send(hello1Bytes, wgpstypes.DataChannel, wgpstypes.SyncRoleAlfie)
	messenger.NewMessenger.Transport.Send(hello2Bytes, wgpstypes.DataChannel, wgpstypes.SyncRoleAlfie)
	messenger.NewMessenger.Transport.Send(hello3Bytes, wgpstypes.DataChannel, wgpstypes.SyncRoleAlfie)
	time.Sleep(time.Second * 2)
}
