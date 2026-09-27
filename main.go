package main

import (
	fit "fit/fit"
)


func main() {
	rootCmd := fit.NewRootCmd()

	fit.InitCLI(rootCmd)
	fit.Execute(rootCmd)
}
