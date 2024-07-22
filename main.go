package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	pinagoladastore "github.com/PES-Innovation-Lab/willow-go/PinaGoladaStore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

var textAScii string = ` 
  ___ ___ _  _   _   ___  ___  _      _   ___   _   
 | _ \_ _| \| | /_\ / __|/ _ \| |    /_\ |   \ /_\  
 |  _/| || .' |/ _ \ (_ | (_) | |__ / _ \| |) / _ \ 
 |_| |___|_|\_/_/ \_\___|\___/|____/_/ \_\___/_/ \_\
`

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

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

	fmt.Println(Blue, textAScii)
	fmt.Println(White, "Type", Yellow, "exit", White, "to escape")
	fmt.Println(" Enter namespace: ", Reset)
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
			fmt.Print(White, "Valid commands:\n\n")
			fmt.Print("new:\t\tUsage: new <namespace>\n\t\tdesc: creates a new namespace\n\n")
			fmt.Print("list:\t\tUsage: list\n\t\tdesc: lists all available namespaces\n\n")
			fmt.Print("enter:\t\tUsage: list\n\t\tdesc: enter into an existing namespace\n\n")
			fmt.Print("help:\t\tUsage: help\n\n")
			fmt.Print("exit:\t\tUsage: exit\n\n", Reset)
		case "new":
			if nameSpaces[input] != 255 {
				fmt.Printf("%sCreating new NameSpaceID %s\n", White, input)
				fmt.Println("Are sure you want to create a new NameSpaceID (y/n)", Reset)
				if !scanner.Scan() {
					break LOOP
				}
				decision := scanner.Text()
				if decision == "y" {
					setNamespace = types.NamespaceId(strings.Split(input, " ")[1])
					NameSpaceInteraction(setNamespace)
				} else if decision == "n" {
					fmt.Println(Red, "Namespace creation canceled. Please enter a valid namespace.", Reset)
				} else {
					fmt.Println(Red, "Invalid input. Please enter 'y' or 'n'.", Reset)
				}
			} else {
				setNamespace = types.NamespaceId(input)
				NameSpaceInteraction(setNamespace)
			}
		case "list":
			if len(nameSpaces) > 0 {
				fmt.Println(White, "available NameSpaces:", Reset)
				for nameSpace := range nameSpaces {
					fmt.Println(nameSpace)
				}
			} else {
				fmt.Println(White, "no namespaces available")
				fmt.Print("use the new command to create a new namespace\n\n")
				fmt.Println("usage: new <namespace>", Reset)

			}
		case "exit":
			fmt.Println(White, "Exiting...", Reset)
			break LOOP
		case "enter":
			if len(objects) != 2 {
				fmt.Println(Red, "invalid usage of command\nusage: enter <namespace>", Reset)
				break
			}
			if nameSpaces[objects[1]] == 255 {
				setNamespace = types.NamespaceId(objects[1])
				NameSpaceInteraction(setNamespace)
			} else {
				fmt.Println(Red, "error: namespace does not exist")
				fmt.Print("use the list command to view available namespaces\n\n")
				fmt.Println("usage: list", Reset)
			}
		default:
			fmt.Println(Red, "invalid command")
			fmt.Println("enter help to list commands!", Reset)
		}

	}
}

func NameSpaceInteraction(namespace types.NamespaceId) {
	WillowStore := pinagoladastore.InitStorage(namespace)
	pinagoladastore.InitKDTree(WillowStore)

LOOPEND:
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			fmt.Println(White, "Exiting Namespace...", Reset)
			break
		}
		objects := strings.Split(input, " ")
		switch objects[0] {
		case "back":
			fmt.Println(textAScii)
			fmt.Println(White, "Type exit to escape")
			fmt.Println("Enter namespace: ", Reset)
			break LOOPEND
		case "help":
			fmt.Println(White, "valid commands:")
			fmt.Println("set:\t\tUsage: set <subspacename> <file/to/path> [<timestamp>]")
			fmt.Println("get:\t\tUsage: get <subspacename> <file/to/path> [<timestamp>]")
			fmt.Println("list:\t\tUsage: list", Reset)
			// fmt.Println("query:\t\tUsage: set <subspacename> <file/to/path> [<timestamp>]")
		case "set":
			if len(objects) < 4 || len(objects) > 5 {
				fmt.Println(Red, "invalid usage of command\nusage: set <subspacename> <path/in/willow> <path/to/file> [<timestamp>]", Reset)
				break
			}
			subSpaceId := []byte(objects[1])
			path := []byte(objects[2])
			insertionFile := objects[3]

			file, err := os.Open(string(insertionFile))
			if err != nil && err.Error() == "open "+string(insertionFile)+": no such file or directory" {
				fmt.Println(Red, "error: file does not exist")
				fmt.Println("please enter a valid path to the file you want to store", err, Reset)
				break
			} else if err != nil {
				fmt.Println(Red, "error opening file:", err, Reset)
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
				fmt.Println(Red, "error reading file:", err, Reset)
				return
			}

			var timestamp uint64
			if len(objects) == 5 {
				timestamp = parseTimeStampToMicroSeconds(objects[4])
				if timestamp == 0 {
					fmt.Println(Red, "invalid timestamp: please enter a valid time", Reset)
					return
				}
			}
			pathBytes := pinagoladastore.ConvertToByteSlices(strings.Split(string(path), "/"))
			prunedEntries, err := WillowStore.Set(
				datamodeltypes.EntryInput{
					Subspace:  subSpaceId,
					Timestamp: timestamp,
					Path:      pathBytes,
					Payload:   payloadBytes,
				},
				subSpaceId,
			)
			if err != nil {
				fmt.Println(Red, "error setting entry:", err, Reset)
				break
			}
			if len(prunedEntries) == 0 {
				fmt.Println(White, "No entries pruned", Reset)
				break
			}
			fmt.Println("Pruned Entries: ")
			for _, entry := range prunedEntries {
				fmt.Printf("%sSubspace: %s, Path: [%s], Timestamp: %d%s\n", White, entry.Subspace_id, makePath(entry.Path), entry.Timestamp, Reset)
			}

		case "get":
			if len(objects) != 3 {
				fmt.Println(Red, "invalid usage of command\nusage: get <subspacename> <file/to/path> [<timestamp>]", Reset)
				break
			}

			subSpaceId := []byte(objects[1])
			path := []byte(objects[2])
			pathBytes := pinagoladastore.ConvertToByteSlices(strings.Split(string(path), "/"))

			var timestamp uint64
			if len(objects) == 4 {
				timestamp = parseTimeStampToMicroSeconds(objects[3])
				if timestamp == 0 {
					fmt.Println(Red, "invalid timestamp", Reset)
					return
				}
			}
			encodedValue, err := WillowStore.EntryDriver.Get(subSpaceId, pathBytes)
			if err != nil {
				fmt.Println(Red, "error getting entry:", err, Reset)
			}
			returnedPayload, err := WillowStore.PayloadDriver.Get(encodedValue.Entry.Payload_digest)
			if err != nil {
				fmt.Println(Red, "error getting payload", err, Reset)
			}

			fmt.Println(string(returnedPayload.Bytes()))

		case "list":
			if len(objects) > 1 {
				fmt.Println(Red, "invalid usage of command\nusage: list", Reset)
			}
			nodes := WillowStore.List()
			sort.Slice(nodes, func(i, j int) bool {
				return utils.OrderBytes(nodes[i].Subspace, nodes[j].Subspace) < 0
			})

			// Print the header of the table
			fmt.Printf("%s %-20s %-20s %-20s\n", White, "Subspace", "Timestamp", "Path")
			fmt.Println(strings.Repeat("-", 60), Reset)

			// Print each node in the sorted list
			for _, node := range nodes {
				fmt.Printf("%s%-20s %-20d %-20s%s\n", White, node.Subspace, node.Timestamp, makePath(node.Path), Reset)
			}

		// case "query":
		case "clear":
			fmt.Println("\003[H\033[2J")
			fmt.Println(Blue, textAScii, Reset)
		default:
			fmt.Println(Red, "invalid command\nenter help to list commands!", Reset)
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

func makePath(path types.Path) string {
	pathStr := string(path[0])
	for i := 1; i < len(path); i++ {
		pathStr += "/" + string(path[i])
	}
	return pathStr
}
