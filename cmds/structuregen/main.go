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
	cppOut := flag.String("cpp-out", "", "Path to generated cpp file")
	cppNamespace := flag.String("cpp-namespace", "", "Namespace name of generated cpp file")

	flag.Parse()

	structureInfoList, err := readSchema(*src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *goOut != "" {
		err = generateGo(structureInfoList, *goOut, *goPackageName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if *cppOut != "" {
		err = generateCpp(structureInfoList, *cppOut, *cppNamespace)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	os.Exit(0)
}

func readSchema(schemaFilePath string) (gen.StructureInfoList, error) {
	schemaFile, err := os.Open(schemaFilePath)
	if err != nil {
		return nil, err
	}
	defer schemaFile.Close()

	return gen.ReadStructureInfoYAML(schemaFile)
}

func generateGo(structureInfoList gen.StructureInfoList, goOut, goPackageName string) error {
	genFile, err := os.Create(goOut)
	if err != nil {
		return err
	}
	defer genFile.Close()

	err = gen.GenGoFile(genFile, goPackageName, structureInfoList)

	return err
}

func generateCpp(structureInfoList gen.StructureInfoList, cppOut, cppNamespace string) error {
	genFile, err := os.Create(cppOut)
	if err != nil {
		return err
	}
	defer genFile.Close()

	err = gen.GenCppFile(genFile, cppNamespace, structureInfoList)

	return err
}
