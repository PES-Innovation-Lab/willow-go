package tests

import (
	"fmt"
	"testing"

	pinagoladastore "github.com/PES-Innovation-Lab/willow-go/PinaGoladaStore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

func TestNewWgpsMessenger(t *testing.T) {
	WillowStore := (*pinagoladastore.InitStorage(types.NamespaceId("myspace")))
	type args[
		ReadCapability any,
		Receiver types.SubspaceId,
		SyncSignature,
		ReceiverSecretKey,
		PsiGroup any,
		PsiScalar int,
		SubspaceCapability any,
		SubspaceReceiver types.SubspaceId,
		SyncSubspaceSignature,
		SubspaceSecretKey any,
		Prefingerprint,
		Fingerprint constraints.Ordered,
		AuthorisationToken,
		StaticToken,
		DynamicToken string,
		AuthorisationOpts []byte,
		K constraints.Unsigned,
	] struct {
		opts             wgps.WgpsMessengerOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]
		newMessengerChan chan wgps.NewMessengerReturn[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]
	}
	tests := []struct {
		name string
		args args[
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
		]
	}{
		{
			name: "basic_test",
			args: args[string, types.SubspaceId, string, string, string, int, string, types.SubspaceId, string, string, string, string, string, string, string, []byte, uint]{
				opts: wgps.WgpsMessengerOpts[string, types.SubspaceId, string, string, string, int, string, types.SubspaceId, string, string, string, string, string, string, string, []byte, uint]{
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
				},
				newMessengerChan: make(chan wgps.NewMessengerReturn[string, types.SubspaceId, string, string, string, int, string, types.SubspaceId, string, string, string, string, string, string, string, []byte, uint], 1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				messenger := <-tt.args.newMessengerChan
				if messenger.Error != nil {
					t.Fatalf("Error in creating messenger: %v", messenger.Error)
				}
				fmt.Println(messenger.NewMessenger)
			}()
			wgps.NewWgpsMessenger(tt.args.opts, tt.args.newMessengerChan)

		})
	}
}
