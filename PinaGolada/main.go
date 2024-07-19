package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	PinaGolada "github.com/PES-Innovation-Lab/willow-go/PinaGolada/Store"
	"github.com/PES-Innovation-Lab/willow-go/types"
)

var pinacoladaAscii string = `
    %#&
      %&&
        #&&
     ,,,,............,,
      ,,,,,,,.........
      ,,,,,,..........
      ,,..............
      ,...............
      ................,
     ...................
    ....................
    ......,,,....,......,
    ..,,,,,,,,,,,,,,**,**
    ...,,,,,*,,,,,,,,,**,
     ,,,,,,************.
       *************/
            /////
             &,(
          , ..*.*...
      ....//  .  .//...
       %,,,,,..,,,,,/#
`

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
	// rootCmd := &cobra.Command{
	// 	Use:   "pinalada",
	// 	Short: "An instantiation of the WillowGo protocol!",
	// 	Long:  "Willow is p2p protocol for file storing and sharing. Pinagolada is an instantiation of the protocll with specific parameters",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		fmt.Println("pina: try 'pina --help' or 'pina - h' for more information")
	// 	},
	// }
	//
	// setCmd := &cobra.Command{
	// 	Use:   "set",
	// 	Short: "Command to enter a new Entry",
	// 	Long:  "Enter the subspace, time(optional) and path of the file which u want to insert into willow",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		fmt.Println(subspaceName)
	// 	},
	// }
	//
	// rootCmd.AddCommand(setCmd)
	// setCmd.Flags().StringVarP(&subspaceName, "sub", "s", "", "Enter subspace name")
	// setCmd.MarkFlagRequired("sub")
	//
	// err := rootCmd.Execute()
	// if err != nil {
	// 	log.Fatal(err)
	// }
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

	NameSpaceInteraction()

	for {
		fmt.Println(pinacoladaAscii)
		fmt.Println(textAScii)
		fmt.Println("Enter namespace: ")
		fmt.Print("> ")

		var cont bool = false
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			fmt.Println("Exiting...")
			break
		} else if strings.Trim(input, "\t\n ") == "" {
			fmt.Println("Please enter a valid namespace")
		} else if nameSpaces[input] != 255 {
			fmt.Printf("Creating new NameSpaceID %s\n", input)
			fmt.Println("Are sure you want to create a new NameSpaceID (yes/no)")
			decision := scanner.Text()
			if decision == "yes" {
				setNamespace = types.NamespaceId(input)
				fmt.Printf("Initiating %s", setNamespace)
				cont = true
				break

			}
		} else {
			setNamespace = types.NamespaceId(input)
			fmt.Printf("Initiating %s", setNamespace)
			cont = true
			break
		}

		if cont {
			NameSpaceInteraction()

		}
		fmt.Println("You entered:", input)
		fmt.Println("exit to escape")
	}
}

func NameSpaceInteraction(namespace types.NamespaceId) {

	WillowStore := PinaGolada.InitStorage(namespace)
	PinaGolada.Init
	fmt.Println(textAScii)
	for {
		fmt.Println()
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
		// Process the input
		objects := strings.Split(input, " ")
		switch objects[0] {
		case "help":
			fmt.Println("Valid commands:")
			fmt.Println("set:\t\tUsage: set <subspacename> <file/to/path> [<timestamp>]")
			fmt.Println("get:\t\tUsage: get <subspacename> <file/to/path> [<timestamp>]")
			fmt.Println("list:\t\tUsage: list")
			fmt.Println("query:\t\tUsage: set <subspacename> <file/to/path> [<timestamp>]")
		case "set":
			if len(objects) < 2 || len(objects) > 3 {
				fmt.Println("invalid usage of command\nusage: set <subspacename> <file/to/path> [<timestamp>]")
			}
		case "get":
			if len(objects) < 2 || len(objects) > 3 {
				fmt.Println("invalid usage of command\nusage: get <subspacename> <file/to/path> [<timestamp>]")
			}
		case "list":
			if len(objects) > 1 {
				fmt.Println("invalid usage of command\nusage: list")
			}
		case "query":
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
}
