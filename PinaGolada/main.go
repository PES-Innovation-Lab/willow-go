package PinaGolada

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pinalada",
		Short: "An instantiation of the WillowGo protocol!",
		Long:  "Willow is p2p protocol for file storing and sharing. Pinagolada is an instantiation of the protocll with specific parameters",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Niggesh")
		},
	}

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
