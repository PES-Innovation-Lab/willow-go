package animation

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kdnode"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/rand"
)

const (
	w      = float64(80)
	startx = float32(400)
	starty = float32(600)
)

func Start_animation() {
	fmt.Println("Hello mother")
	values := []kdnode.Key{
		{Timestamp: 1704067200, Subspace: types.SubspaceId("Manas"), Path: types.Path{[]byte("hello"), []byte("bye")}, Fingerprint: "fingerprint1"},
		{Timestamp: 1729742700, Subspace: types.SubspaceId("Samar"), Path: types.Path{[]byte("hello"), []byte("1234"), []byte("testing")}, Fingerprint: "fingerprint2"},
		{Timestamp: 1704067199, Subspace: types.SubspaceId("Samarth"), Path: types.Path{[]byte("test"), []byte("path"), []byte("hello")}, Fingerprint: "fingerprint3"},
	}

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
	fmt.Println("Goodmorning")
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
	// var values []kdnode.Key

	for !rl.WindowShouldClose() {
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

func draw_files(entries []kdnode.Key, subspace_map map[string]int, time_map map[uint64]int, path_map map[string]int, textures []rl.Texture2D) {
	fmt.Println(subspace_map, path_map)
	for _, entry := range entries {
		uind := subspace_map[string(entry.Subspace)]
		tind := time_map[entry.Timestamp]
		pind := path_map[makePath(entry.Path)]

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

func makePath(path types.Path) string {
	pathStr := string(path[0])
	for i := 1; i < len(path); i++ {
		pathStr += "/" + string(path[i])
	}
	return pathStr
}

// convert_timestamp converts a uint64 Unix timestamp to a readable string.
func convert_timestamp(timestamp uint64) string {
	ts := int64(timestamp)
	t := time.Unix(ts, 0)
	readable_time := t.Format("2006-01-02 15:04:05")

	return readable_time
}
