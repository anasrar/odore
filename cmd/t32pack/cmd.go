package main

import (
	"log"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "t32pack",
	Short: "Pack PNG to T32",
	Run: func(cmd *cobra.Command, args []string) {
		if Input != "" && !IsGUI {
			if err := pack(Input); err != nil {
				log.Fatal(err)
			}
			return
		}

		if Input != "" && IsGUI {
			if err := gui(Input); err != nil {
				log.Fatal(err)
			}
			return
		}

		if err := gui(Input); err != nil {
			log.Fatal(err)
		}
	},
}
