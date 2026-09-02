package main

import (
	"log"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "t32unpack",
	Short: "Unpack T32 to PNG",
	Run: func(cmd *cobra.Command, args []string) {
		if Input != "" && !IsGUI {
			if err := unpack(Input, Offset); err != nil {
				log.Fatal(err)
			}
			return
		}

		if Input != "" && IsGUI {
			if err := gui(Input, Offset); err != nil {
				log.Fatal(err)
			}
			return
		}

		if err := gui(Input, Offset); err != nil {
			log.Fatal(err)
		}
	},
}
