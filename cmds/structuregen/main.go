package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/y-akahori-ramen/ueLogHandler/gen"
)

func main() {
	src := flag.String("src", "", "Path to structure schema file")
	goOut := flag.String("go-out", "", "Path to generated go file")
	goPackageName := flag.String("go-package", "", "Package name of generated go file")

	flag.Parse()

	err := generate(*src, *goOut, *goPackageName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func generate(src, goOut, goPackageName string) error {
	schemaFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer schemaFile.Close()

	structureInfoList, err := gen.ReadStructureInfoYAML(schemaFile)
	if err != nil {
		return err
	}

	genFile, err := os.Create(goOut)
	if err != nil {
		return err
	}
	defer genFile.Close()

	err = gen.GenGoFile(genFile, goPackageName, structureInfoList)

	return err
}
