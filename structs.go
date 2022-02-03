package main

type Args struct {
	OutPath string `arg:"-o" help:"Where to write credentials to. File extension must be \".txt\" or \".json\". Path will be made if it doesn't already exist."`
	Print   bool   `arg:"-p" help:"Write JSON to stdout."`
}
