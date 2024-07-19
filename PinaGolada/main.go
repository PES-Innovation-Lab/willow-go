package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	entrydriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/entry_driver"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/cockroachdb/pebble"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "storecli",
		Short: "A CLI for interacting with the store",
	}

	chooseNamespaceCmd = &cobra.Command{
		Use:   "choose-namespace",
		Short: "Choose a namespace",
		Run:   chooseNamespace,
	}

	setPayloadCmd = &cobra.Command{
		Use:   "set-payload",
		Short: "Set a payload in the store",
		Run:   setPayload,
	}

	getPayloadCmd = &cobra.Command{
		Use:   "get-payload",
		Short: "Get a payload from the store",
		Run:   getPayload,
	}

	namespaces = []string{"namespace1", "namespace2", "namespace3"}
	Storage    *store.Store[uint64, uint64, uint8, []byte, string]
	namespace  types.NamespaceId
)

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(chooseNamespaceCmd)
	rootCmd.AddCommand(setPayloadCmd)
	rootCmd.AddCommand(getPayloadCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}

func chooseNamespace(cmd *cobra.Command, args []string) {
	fmt.Println("Available namespaces:")
	for i, ns := range namespaces {
		fmt.Printf("%d: %s\n", i+1, ns)
	}

	var choice int
	fmt.Print("Choose a namespace (enter number): ")
	fmt.Scan(&choice)

	if choice < 1 || choice > len(namespaces) {
		log.Fatal("Invalid choice")
	}

	namespace = types.NamespaceId(namespaces[choice-1])
	Storage = InitStorage(namespace)
	fmt.Printf("Namespace '%s' selected and store initialized.\n", namespaces[choice-1])
}

func convertToByteSlices(strings []string) types.Path {
	byteSlices := make([][]byte, len(strings))
	for i, str := range strings {
		byteSlices[i] = []byte(str)
	}
	return byteSlices
}
func setPayload(cmd *cobra.Command, args []string) {
	if Storage == nil {
		log.Fatal("Store is not initialized. Please choose a namespace first.")
	}

	var subspace, payload, path string
	fmt.Print("Enter subspace: ")
	fmt.Scan(&subspace)
	fmt.Print("Enter payload: ")
	fmt.Scan(&payload)
	fmt.Print("Enter path (/ separated): ")
	fmt.Scan(&path)
	entrypath := convertToByteSlices(strings.Split(path, "/"))

	entryInput := datamodeltypes.EntryInput{
		Subspace:  []byte(subspace),
		Payload:   []byte(payload),
		Timestamp: uint64(time.Now().UnixMicro()),
		Path:      entrypath,
	}

	authOpts := []byte(subspace)
	prunedEntries := Storage.Set(entryInput, authOpts)

	fmt.Println("Payload set in the store.")
	fmt.Printf("Pruned entries: %v\n", prunedEntries)
}

func getPayload(cmd *cobra.Command, args []string) {
	if Storage == nil {
		log.Fatal("Store is not initialized. Please choose a namespace first.")
	}

	var subspace, path string
	fmt.Print("Enter subspace: ")
	fmt.Scan(&subspace)
	fmt.Print("Enter path (comma separated): ")
	fmt.Scan(&path)
	entrypath := convertToByteSlices(strings.Split(path, "/"))

	position := types.Position3d{
		Subspace: []byte(subspace),
		Path:     entrypath,
	}

	payload := Storage.GetPayload(position)
	fmt.Printf("Payload: %s\n", payload)
}

// The InitStorage function remains the same as provided by you earlier
func InitStorage(nameSpaceId types.NamespaceId) *store.Store[uint64, uint64, uint8, []byte, string] {

	payloadRefDb, err := pebble.Open("willow/payloadrefcounter", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}

	payloadRefKVstore := kv_driver.KvDriver{Db: payloadRefDb}
	PayloadReferenceCounter := payloadDriver.PayloadReferenceCounter{
		Store: payloadRefKVstore,
	}

	entryDb, err := pebble.Open("willow/entries", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	entryKvStore := kv_driver.KvDriver{Db: entryDb}

	PayloadLock := &sync.Mutex{}
	TestPayloadDriver := payloadDriver.MakePayloadDriver("willow/payload", store.TestPayloadScheme, PayloadLock)

	entryDriver := entrydriver.EntryDriver[uint64, uint64, uint8]{
		PayloadReferenceCounter: PayloadReferenceCounter,
		Opts: struct {
			KVDriver          kv_driver.KvDriver
			NamespaceScheme   datamodeltypes.NamespaceScheme
			SubspaceScheme    datamodeltypes.SubspaceScheme
			PayloadScheme     datamodeltypes.PayloadScheme
			PathParams        types.PathParams[uint8]
			FingerprintScheme datamodeltypes.FingerprintScheme[uint64, uint64]
		}{
			KVDriver:          entryKvStore,
			NamespaceScheme:   store.TestNameSpaceScheme,
			SubspaceScheme:    store.TestSubspaceScheme,
			PayloadScheme:     store.TestPayloadScheme,
			PathParams:        store.TestPathParams,
			FingerprintScheme: store.TestFingerprintScheme,
		},
	}
	TestPrefixDriver := kv_driver.PrefixDriver[uint8]{}

	return &store.Store[uint64, uint64, uint8, []byte, string]{
		Schemes:            store.StoreSchemes,
		EntryDriver:        entryDriver,
		PayloadDriver:      TestPayloadDriver,
		NameSpaceId:        nameSpaceId,
		IngestionMutexLock: sync.Mutex{},
		PrefixDriver:       TestPrefixDriver,
	}
}
