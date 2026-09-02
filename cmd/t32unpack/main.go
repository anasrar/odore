package main

func init() {
	cmd.Flags().StringVarP(&Input, "input", "i", "", "input file")
	cmd.MarkFlagRequired("input")
	cmd.Flags().Uint32VarP(&Offset, "offset", "o", 0x0, "offset from input file")
}

func main() {
	cmd.Execute()
}
