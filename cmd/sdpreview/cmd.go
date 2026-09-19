package main

import (
	"log"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "ppkviewer",
	Short: "View 3D model on PPK",
	Run: func(cmd *cobra.Command, args []string) {
		if err := gui(Input); err != nil {
			log.Fatal(err)
		}
	},
}
