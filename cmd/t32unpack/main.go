package main

func init() {
	cmd.Flags().StringVarP(&Input, "input", "i", "", "input file")
	cmd.Flags().Uint32VarP(&Offset, "offset", "o", 0x0, "offset from input file")
	cmd.Flags().BoolVarP(&IsGUI, "gui", "g", false, "open GUI")
}

func main() {
	cmd.Execute()
}
