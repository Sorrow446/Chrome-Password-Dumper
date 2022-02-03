package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"main/linux"
	"main/utils"
	"main/windows"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alexflint/go-arg"
)

func parseArgs() (*Args, bool, error) {
	var (
		args    Args
		jsonOut bool
	)
	arg.MustParse(&args)
	if args.OutPath == "" && !args.Print {
		return nil, false, errors.New("Output or format arg required.")
	}
	if args.OutPath != "" {
		if strings.HasSuffix(args.OutPath, ".json") {
			jsonOut = true
		} else if strings.HasSuffix(args.OutPath, ".txt") {
			jsonOut = false
		} else {
			return nil, false, errors.New(`Invalid output format. File extension must be ".txt" or ".json".`)
		}
	}
	return &args, jsonOut, nil
}

func makeDirs(pathWithFname string) error {
	path := filepath.Dir(pathWithFname)
	if path != "." {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func marshal(loginData *[]utils.LoginData) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(loginData)
	if err != nil {
		return nil, err
	}
	bufBytes := buf.Bytes()
	return bufBytes[:len(bufBytes)-1], nil
}

func writeJson(m []byte, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(m)
	return err
}

func writePlain(loginData *[]utils.LoginData, outPath string) error {
	total := len(*loginData)
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	for i, d := range *loginData {
		line := fmt.Sprintf(
			"URL: %s\nUsername: %s\nPassword: %s\n", d.URL, d.Username, d.Password)
		if i+1 != total {
			line += "\n"
		}
		_, err = f.WriteString(line)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var (
		loginData *[]utils.LoginData
		err       error
		m         []byte
	)
	args, jsonOut, err := parseArgs()
	if err != nil {
		panic(err)
	}
	if args.OutPath != "" {
		err = makeDirs(args.OutPath)
		if err != nil {
			panic(err)
		}
	}
	switch runtime.GOOS {
	case "linux":
		loginData, err = linux.Run()
	case "windows":
		loginData, err = windows.Run()
	default:
		panic("Unsupported OS.")
	}
	if err != nil {
		panic(err)
	}
	if jsonOut || args.Print {
		m, err = marshal(loginData)
		if err != nil {
			panic(err)
		}
		if args.Print {
			fmt.Print(string(m))
		}
	}
	if args.OutPath != "" {
		if jsonOut {
			err = writeJson(m, args.OutPath)
			if err != nil {
				panic(err)
			}
		} else {
			err = writePlain(loginData, args.OutPath)
			if err != nil {
				panic(err)
			}
		}
	}
}
