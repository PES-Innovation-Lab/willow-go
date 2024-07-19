package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	pinagoladastore "github.com/PES-Innovation-Lab/willow-go/PinaGoladaStore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
)

var textAScii string = `
░▒▓███████▓▒░░▒▓█▓▒░▒▓███████▓▒░ ░▒▓██████▓▒░ ░▒▓██████▓▒░ ░▒▓██████▓▒░░▒▓█▓▒░       ░▒▓██████▓▒░░▒▓███████▓▒░ ░▒▓██████▓▒░  
░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓███████▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓████████▓▒░▒▓█▓▒▒▓███▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓████████▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓████████▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░ ░▒▓██████▓▒░░▒▓████████▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓███████▓▒░░▒▓█▓▒░░▒▓█▓▒░ 
                                                                                                                             `

var setNamespace types.NamespaceId

var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)

func main() {
	fmt.Println("\033[H\033[2J")
	dir := "willow"
	f, err := os.Open(dir)
	nameSpaces := make(map[string]uint8)
	if err != nil && strings.Compare(err.Error(), "open willow: no such file or directory") != 0 {
		fmt.Println(err)
		return
	} else if err == nil {
		files, err := f.Readdir(-1)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, file := range files {
			if file.IsDir() {
				nameSpaces[file.Name()] = 255
			}
		}
	}
	defer f.Close()

	fmt.Println(textAScii)
	fmt.Println("Enter namespace: ")
	fmt.Println("exit to escape")
LOOP:
	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break LOOP
		}
		input := scanner.Text()
		objects := strings.Split(input, " ")
		switch objects[0] {
		case "help":
			fmt.Print("Valid commands:\n\n")
			fmt.Print("new:\t\tUsage: new <namespace>\n\t\tdesc: creates a new namespace\n\n")
			fmt.Print("list:\t\tUsage: list\n\t\tdesc: lists all available namespaces\n\n")
			fmt.Print("help:\t\tUsage: help\n\n")
			fmt.Print("exit:\t\tUsage: exit\n\n")
		case "new":
			if nameSpaces[input] != 255 {
				fmt.Printf("Creating new NameSpaceID %s\n", input)
				fmt.Println("Are sure you want to create a new NameSpaceID (y/n)")
				if !scanner.Scan() {
					break LOOP
				}
				decision := scanner.Text()
				if decision == "y" {
					setNamespace = types.NamespaceId(strings.Split(input, " ")[1])
					NameSpaceInteraction(setNamespace)
				} else if decision == "n" {
					fmt.Println("Namespace creation canceled. Please enter a valid namespace.")
				} else {
					fmt.Println("Invalid input. Please enter 'y' or 'n'.")
				}
			} else {
				setNamespace = types.NamespaceId(input)
				NameSpaceInteraction(setNamespace)
			}
		case "list":
			if len(nameSpaces) > 0 {
				fmt.Println("available NameSpaces:")
				for nameSpace := range nameSpaces {
					fmt.Println(nameSpace)
				}
			} else {
				fmt.Println("no namespaces available")
				fmt.Print("use the new command to create a new namespace\n\n")
				fmt.Println("usage: new <namespace>")

			}
		case "exit":
			fmt.Println("Exiting...")
			break LOOP
		case "enter":
			if len(objects) != 2 {
				fmt.Println("invalid usage of command\nusage: enter <namespace>")
				break
			}
			if nameSpaces[objects[1]] == 255 {
				setNamespace = types.NamespaceId(objects[1])
				NameSpaceInteraction(setNamespace)
			} else {
				fmt.Println("error: namespace does not exist")
				fmt.Print("use the list command to view available namespaces\n\n")
				fmt.Println("usage: list")
			}
		default:
			fmt.Println("invalid command")
			fmt.Println("enter help to list commands!")
		}

	}
}

func NameSpaceInteraction(namespace types.NamespaceId) {
	WillowStore := pinagoladastore.InitStorage(namespace)
	pinagoladastore.InitKDTree(WillowStore)
	fmt.Println("\003[H\033[2J")
	fmt.Println(textAScii)
LOOPEND:
	for {
		fmt.Println("type 'exit' to quit:")
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			fmt.Println("Exiting...")
			break
		}
		objects := strings.Split(input, " ")
		switch objects[0] {
		case "back":
			fmt.Println(textAScii)
			fmt.Println("enter namespace: ")
			fmt.Println("exit to escape")
			break LOOPEND
		case "help":
			fmt.Println("valid commands:")
			fmt.Println("set:\t\tUsage: set <subspacename> <file/to/path> [<timestamp>]")
			fmt.Println("get:\t\tUsage: get <subspacename> <file/to/path> [<timestamp>]")
			fmt.Println("list:\t\tUsage: list")
			// fmt.Println("query:\t\tUsage: set <subspacename> <file/to/path> [<timestamp>]")
		case "set":
			if len(objects) < 4 || len(objects) > 5 {
				fmt.Println("invalid usage of command\nusage: set <subspacename> <path/in/willow> <path/to/file> [<timestamp>]")
				break
			}
			subSpaceId := []byte(objects[1])
			path := []byte(objects[2])
			insertionFile := objects[3]

			file, err := os.Open(string(insertionFile))
			if err != nil && err.Error() == "open "+string(insertionFile)+": no such file or directory" {
				fmt.Println("error: file does not exist")
				fmt.Println("please enter a valid path to the file you want to store", err)
				break
			} else if err != nil {
				fmt.Println("error opening file:", err)
				break
			}
			defer file.Close()
			fileInfo, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}
			fileSize := fileInfo.Size()
			payloadBytes := make([]byte, fileSize)
			_, err = io.ReadFull(file, payloadBytes)
			if err != nil {
				fmt.Println("error reading file:", err)
				return
			}

			var timestamp uint64
			if len(objects) == 5 {
				timestamp = parseTimeStampToMicroSeconds(objects[4])
				if timestamp == 0 {
					fmt.Println("invalid timestamp: please enter a valid time")
					return
				}
			}
			pathBytes := pinagoladastore.ConvertToByteSlices(strings.Split(string(path), "/"))
			prunedEntries := WillowStore.Set(
				datamodeltypes.EntryInput{
					Subspace:  subSpaceId,
					Timestamp: timestamp,
					Path:      pathBytes,
					Payload:   payloadBytes,
				},
				subSpaceId,
			)
			if len(prunedEntries) == 0 {
				fmt.Println("No entries pruned")
				break
			}
			fmt.Println("Pruned Entries: ")
			for _, entry := range prunedEntries {
				fmt.Printf("Subspace: %s, Path: %s, Timestamp: %d\n", entry.Subspace_id, entry.Path, entry.Timestamp)
			}

		case "get":
			if len(objects) != 3 {
				fmt.Println("invalid usage of command\nusage: get <subspacename> <file/to/path> [<timestamp>]")
				break
			}

			subSpaceId := []byte(objects[1])
			path := []byte(objects[2])
			pathBytes := pinagoladastore.ConvertToByteSlices(strings.Split(string(path), "/"))

			var timestamp uint64
			if len(objects) == 4 {
				timestamp = parseTimeStampToMicroSeconds(objects[3])
				if timestamp == 0 {
					fmt.Println("invalid timestamp")
					return
				}
			}
			encodedValue, err := WillowStore.EntryDriver.Get(subSpaceId, pathBytes)
			fmt.Println(encodedValue)
			if err != nil {
				log.Fatal(err)
			}
			returnedPayload, err := WillowStore.PayloadDriver.Get(encodedValue.Entry.Payload_digest)
			fmt.Println(returnedPayload)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(returnedPayload.Bytes()))

		case "list":
			if len(objects) > 1 {
				fmt.Println("invalid usage of command\nusage: list")
			}

		// case "query":
		case "clear":
			fmt.Println("\003[H\033[2J")
			fmt.Println(textAScii)
		default:
			fmt.Println("invalid command\nenter help to list commands!")
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
		}
	}
	WillowStore.EntryDriver.Opts.KVDriver.Close()
	WillowStore.EntryDriver.PayloadReferenceCounter.Store.Close()
}

func parseTimeStampToMicroSeconds(timestamp string) uint64 {
	dateTime := strings.Split(timestamp, ":")
	if len(dateTime) != 2 {
		return 0
	}
	hours, err := strconv.ParseUint(dateTime[0], 10, 64)
	if err != nil {
		return 0
	}
	minutes, err := strconv.ParseUint(dateTime[1], 10, 64)
	if err != nil {
		return 0
	}
	return (hours * 60 * 60 * 1000000) + (minutes * 60 * 1000000)
}
