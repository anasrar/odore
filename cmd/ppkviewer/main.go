package main

func init() {
	cmd.Flags().StringVarP(&Input, "input", "i", "", "input file")
}

func main() {
	cmd.Execute()
}
