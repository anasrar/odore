package main

func init() {
	cmd.Flags().StringVarP(&Input, "input", "i", "", "Input file")
	cmd.Flags().BoolVarP(&IsGUI, "gui", "g", false, "open GUI")
}

func main() {
	cmd.Execute()
}
