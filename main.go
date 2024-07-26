package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	pinagoladastore "github.com/PES-Innovation-Lab/willow-go/PinaGoladaStore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kdnode"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/rand"
)

var textAScii string = ` 
  ___ ___ _  _   _   ___  ___  _      _   ___   _   
 | _ \_ _| \| | /_\ / __|/ _ \| |    /_\ |   \ /_\  
 |  _/| || .' |/ _ \ (_ | (_) | |__ / _ \| |) / _ \ 
 |_| |___|_|\_/_/ \_\___|\___/|____/_/ \_\___/_/ \_\
`

const (
	w      = float64(80)
	startx = float32(400)
	starty = float32(600)
)

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
				fmt.Printf("%sCreating new NameSpaceID %s\n", White, objects[1])
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
	// fmt.Println("Hello father")
	// animationValues := make(chan []kdnode.Key)
	// fmt.Println("Created channel")
	// go animation.Start_animation(animationValues)
LOOPEND:
	for {

		// animationValues <- WillowStore.List()
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			fmt.Println(White, "Exiting...", Reset)
			break
		}
		input = strings.TrimSpace(input)
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

func StartAnimation(valuesChannel chan []kdnode.Key) {
	// values := []kdnode.Key{
	// 	{Timestamp: 1704067200, Subspace: types.SubspaceId("Manas"), Path: types.Path{[]byte("hello"), []byte("bye")}, Fingerprint: "fingerprint1"},
	// 	{Timestamp: 1729742700, Subspace: types.SubspaceId("Samar"), Path: types.Path{[]byte("hello"), []byte("1234"), []byte("testing")}, Fingerprint: "fingerprint2"},
	// 	{Timestamp: 1704067199, Subspace: types.SubspaceId("Samarth"), Path: types.Path{[]byte("test"), []byte("path"), []byte("hello")}, Fingerprint: "fingerprint3"},
	// }

	rl.InitWindow(1200, 900, "learning this bad boy yeeyee")
	defer rl.CloseWindow()

	texture1 := rl.LoadTexture(`G:\PIL\main_project\raylib\images\img1.png`)
	texture2 := rl.LoadTexture(`G:\PIL\main_project\raylib\images\img2.png`)
	texture3 := rl.LoadTexture(`G:\PIL\main_project\raylib\images\img3.png`)
	texture4 := rl.LoadTexture(`G:\PIL\main_project\raylib\images\img4.png`)
	defer rl.UnloadTexture(texture1)
	defer rl.UnloadTexture(texture2)
	defer rl.UnloadTexture(texture3)
	defer rl.UnloadTexture(texture4)

	textures := []rl.Texture2D{texture1, texture2, texture3, texture4}

	rl.SetTargetFPS(60)

	// var paths []string

	// paths := []string{
	// 	"/usr/local/bin",
	// 	"/home/user/documents",
	// 	"/var/log/apache2",
	// 	"/etc/nginx/sites-available",
	// 	"/tmp/cache/session",
	// }

	// times := []uint64{
	// 	1704067200,
	// 	1729742700,
	// 	1704067199,
	// 	1704067199,
	// 	1704067199,
	// }

	// payload_count := 5

	// h1 := int32(path_count * 20)
	// h2 := int32(time_count * 20)
	var values []kdnode.Key
	temp := <-valuesChannel

	for !rl.WindowShouldClose() {
		if temp != nil {
			values = temp
		}
		var sortedSubspaceArray []types.SubspaceId

		sortedSubspaceMap := func() map[string]int {
			sort.Slice(values, func(i, j int) bool {
				return utils.OrderSubspace(values[i].Subspace, values[j].Subspace) == -1
			})
			subspaceMap := make(map[string]int)

			for i, value := range values {
				sortedSubspaceArray = append(sortedSubspaceArray, value.Subspace)
				subspaceMap[string(value.Subspace)] = i
			}

			return subspaceMap
		}()

		var sortedPathArray []string
		sortedPathMap := func() map[string]int {
			sort.Slice(values, func(i, j int) bool {
				return utils.OrderPath(values[i].Path, values[j].Path) == -1
			})
			pathMap := make(map[string]int)

			for i, value := range values {
				sortedPathArray = append(sortedPathArray, makePath(value.Path))
				pathMap[makePath(value.Path)] = i
			}
			return pathMap
		}()
		var sortedTimeArray []uint64
		sortedTimeMap := func() map[uint64]int {
			sort.Slice(values, func(i, j int) bool {
				return values[i].Timestamp < values[j].Timestamp
			})
			timeMap := make(map[uint64]int)

			for i, value := range values {
				sortedTimeArray = append(sortedTimeArray, value.Timestamp)
				timeMap[value.Timestamp] = i
			}

			return timeMap
		}()
		fmt.Println(sortedPathMap, sortedSubspaceMap, sortedTimeMap)

		paths := sortedPathArray
		times := sortedTimeArray

		user_count := len(sortedSubspaceArray)
		path_count := len(paths)
		time_count := len(times)

		rl.BeginDrawing()

		rl.ClearBackground(rl.White)
		// rl.DrawText("Data Model Animation coming soon", 190, 200, 20, rl.Black)
		// draw_plgm_2(rl.Vector2{X: 100, Y: 100}, 4, rl.Black)
		// draw_plgm_1(rl.Vector2{X: 100, Y: 100}, 4, rl.Blue)
		// draw_plgm_2(rl.Vector2{X: 400, Y: 400}, 4, rl.Black)
		// draw_plgm_1(rl.Vector2{X: 400, Y: 400}, 6, rl.Green)

		for ind, path := range paths {
			draw_right_aligned_text(path, int32(startx)-10, int32(int(starty)-ind*int(w)-45), 18, rl.Black)
		}

		// for ind, time := range times {
		// 	draw_time(time)
		// }

		points := get_starting_points(user_count)
		for ind, point := range points {
			if ind%2 == 0 {
				draw_plgm_1(point, path_count, rl.Gray)
				draw_plgm_2(point, time_count, rl.LightGray)
			} else {
				draw_plgm_1(point, path_count, rl.LightGray)
				draw_plgm_2(point, time_count, rl.Gray)
			}
		}

		draw_times(time_count, times)

		draw_files(values, sortedSubspaceMap, sortedTimeMap, sortedPathMap, textures)

		rl.EndDrawing()
	}
}

func OrderSubspace(a, b types.SubspaceId) int {
	if hex.EncodeToString(a[:]) < hex.EncodeToString(b[:]) {
		return -1
	} else if hex.EncodeToString(a[:]) > hex.EncodeToString(b[:]) {
		return 1
	}
	return 0
}

func OrderPath(a, b types.Path) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		order := OrderBytes(a[i], b[i])
		if order != 0 {
			return order
		}
	}

	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return 1
	}
	return 0
}

func OrderBytes(a, b []byte) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}

	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return 1
	}

	return 0
}

func draw_files(entries []kdnode.Key, subspace_map map[string]int, time_map map[uint64]int, path_map map[string]int, textures []rl.Texture2D) {
	fmt.Println(subspace_map, path_map)
	for _, entry := range entries {
		uind := subspace_map[string(entry.Subspace)]
		tind := time_map[entry.Timestamp]
		pind := path_map[makePath(entry.Path)]
		fmt.Println(makePath(entry.Path), uind, tind, pind)
		draw_file(float64(tind), float64(pind), float64(uind), textures)
	}
}

func draw_file(time_index, path_index, user_index float64, textures []rl.Texture2D) {
	rand.Seed(uint64(time.Now().UnixNano()))
	random_number := rand.Intn(4) + 1
	texture := textures[random_number-1]

	udelx := float32(w * math.Cos(math.Pi/6) * user_index)
	udely := -float32(w * math.Sin(math.Pi/6) * user_index)
	ucentererx := float32((w / 2) * math.Cos(math.Pi/6) * user_index)
	ucenterery := -float32((w / 2) * math.Sin(math.Pi/6) * user_index)
	pdely := -float32(path_index*w) - float32(w/2)
	tdelx := float32(time_index * w * math.Sin(math.Pi/3))
	tdely := float32(time_index * w * math.Cos(math.Pi/3))
	tcentererx := float32((w / 2) * math.Sin(math.Pi/3))
	tcenterery := float32((w / 2) * math.Cos(math.Pi/3))

	// rl.DrawTextureV(texture, rl.Vector2{
	// 	X: startx + udelx + ucentererx + tdelx + tcentererx,
	// 	Y: starty + udely + ucenterery + tdely + tcenterery + pdely,
	// }, rl.White)

	maxWidth, maxHeight := 75.0, 75.0
	aspectRatio := float32(texture.Width) / float32(texture.Height)
	newWidth, newHeight := maxWidth, maxHeight

	if aspectRatio > 1 {
		newHeight = maxWidth / float64(aspectRatio)
	} else {
		newWidth = maxHeight * float64(aspectRatio)
	}

	// Define source and destination rectangles
	sourceRec := rl.NewRectangle(0, 0, float32(texture.Width), float32(texture.Height))
	destRec := rl.NewRectangle(
		startx+udelx+ucentererx+tdelx+tcentererx,
		starty+udely+ucenterery+tdely+tcenterery+pdely,
		float32(newWidth),
		float32(newHeight),
	)
	origin := rl.NewVector2(0, 0)

	// Draw the texture
	rl.DrawTexturePro(texture, sourceRec, destRec, origin, 0, rl.White)
}

func draw_times(time_count int, times []uint64) {
	// time_pos := make([]rl.Vector2, time_count)
	j := math.Sin(math.Pi / 3)
	k := math.Cos(math.Pi / 3)

	for i := 1; i < time_count+1; i++ {
		length := float64(i * int(w))
		x := float32(length * j)
		y := float32(length * k)

		// time_pos[i-1] = rl.Vector2{
		// 	X: startx + x - float32(w/2),
		// 	Y: starty + y - float32(w/2),
		// }
		// fmt.Println(int32(startx + x - float32(w/2)))
		// fmt.Println(int32(starty + y - float32(w/2)))

		// rl.DrawCircleV(rl.Vector2{
		// 	X: startx + x - float32(w/4) - float32((w/4)*3),
		// 	Y: starty + y - float32(w/4),
		// }, 3, rl.Red)

		dash_length := float32(7)
		dash_start := rl.Vector2{
			X: startx + x - float32(w/4) - float32((w / 7.5)),
			Y: starty + y - float32(w/4) - dash_length/2,
		}
		dash_end := rl.Vector2{
			X: dash_start.X,
			Y: dash_start.Y + dash_length,
		}

		// fmt.Printf("Dash %d: start = (%f, %f), end = (%f, %f)\n", i, dash_start.X, dash_start.Y, dash_end.X, dash_end.Y)
		// rl.DrawLineV(dash_start, dash_end, rl.Red)
		rl.DrawLineEx(dash_start, dash_end, 2.5, rl.Black)

		draw_right_aligned_text(convert_timestamp(times[i-1]),
			int32(startx+x-float32(w/8)-float32(w/4)),
			int32(starty+y-float32(w/8)),
			16,
			rl.Black,
		)
	}
}

func draw_right_aligned_text(text string, rightEdgeX, posY, fontSize int32, color rl.Color) {
	textWidth := rl.MeasureText(text, fontSize)
	posX := rightEdgeX - textWidth
	rl.DrawText(text, posX, posY, fontSize, color)
}

func get_starting_points(num int) []rl.Vector2 {
	lis := make([]rl.Vector2, num)

	start := rl.Vector2{
		X: float32(startx),
		Y: float32(starty),
	}

	lis[0] = start
	for i := 1; i < num; i++ {
		lis[i] = rl.Vector2{
			X: lis[i-1].X + float32(w*math.Cos(math.Pi/6)),
			Y: lis[i-1].Y - float32(w*math.Sin(math.Pi/6)),
		}
	}

	return lis
}

func draw_plgm_1(v1 rl.Vector2, path_count int, col rl.Color) {
	v2 := rl.Vector2{
		X: v1.X + float32(w*math.Cos(math.Pi/6)),
		Y: v1.Y - float32(w*math.Sin(math.Pi/6)),
	}

	height := float32(path_count * int(w))

	v3 := rl.Vector2{
		X: v2.X,
		Y: v2.Y - height,
	}

	v4 := rl.Vector2{
		X: v1.X,
		Y: v1.Y - height,
	}

	// println("v1:", v1.X, v1.Y)
	// println("v2:", v2.X, v2.Y)
	// println("v3:", v3.X, v3.Y)
	// println("v4:", v4.X, v4.Y)

	// Draw points to visualize them
	// rl.DrawCircleV(v1, 3, rl.Red)
	// rl.DrawCircleV(v2, 3, rl.Green)
	// rl.DrawCircleV(v3, 3, rl.Blue)
	// rl.DrawCircleV(v4, 3, rl.Yellow)

	// // Draw lines connecting the points
	// rl.DrawLineV(v1, v2, rl.Red)
	// rl.DrawLineV(v2, v3, rl.Green)
	// rl.DrawLineV(v3, v4, rl.Blue)
	// rl.DrawLineV(v4, v1, rl.Yellow)

	rl.DrawTriangle(v1, v2, v3, col)
	rl.DrawTriangle(v1, v3, v4, col)

	// draw borders :)
	border_color := rl.Black
	border_thickness := float32(2)

	rl.DrawLineEx(v1, v2, border_thickness, border_color)
	rl.DrawLineEx(v2, v3, border_thickness, border_color)
	rl.DrawLineEx(v3, v4, border_thickness, border_color)
	rl.DrawLineEx(v4, v1, border_thickness, border_color)
}

func draw_plgm_2(v1 rl.Vector2, time_count int, col rl.Color) {
	length := float64(time_count * int(w))
	x1 := float32(w * math.Cos(math.Pi/6))
	y1 := float32(w * math.Sin(math.Pi/6))

	v2 := rl.Vector2{
		X: v1.X + x1,
		Y: v1.Y - y1,
	}

	x2 := float32(length * math.Sin(math.Pi/3))
	y2 := float32(length * math.Cos(math.Pi/3))

	v3 := rl.Vector2{
		X: v2.X + x2,
		Y: v2.Y + y2,
	}

	v4 := rl.Vector2{
		X: v1.X + x2,
		Y: v1.Y + y2,
	}

	// Draw points to visualize them
	// rl.DrawCircleV(v1, 3, rl.Red)
	// rl.DrawCircleV(v2, 3, rl.Green)
	// rl.DrawCircleV(v3, 3, rl.Blue)
	// rl.DrawCircleV(v4, 3, rl.Yellow)

	// // Draw lines connecting the points
	// rl.DrawLineV(v1, v2, rl.Red)
	// rl.DrawLineV(v2, v3, rl.Green)
	// rl.DrawLineV(v3, v4, rl.Blue)
	// rl.DrawLineV(v4, v1, rl.Yellow)

	rl.DrawTriangle(v3, v2, v1, col)
	rl.DrawTriangle(v3, v1, v4, col)

	// draw borders :)
	border_color := rl.Black
	border_thickness := float32(2)

	rl.DrawLineEx(v1, v2, border_thickness, border_color)
	rl.DrawLineEx(v2, v3, border_thickness, border_color)
	rl.DrawLineEx(v3, v4, border_thickness, border_color)
	rl.DrawLineEx(v4, v1, border_thickness, border_color)
}

// convert_timestamp converts a uint64 Unix timestamp to a readable string.
func convert_timestamp(timestamp uint64) string {
	ts := int64(timestamp)
	t := time.Unix(ts, 0)
	readable_time := t.Format("2006-01-02 15:04:05")

	return readable_time
}
