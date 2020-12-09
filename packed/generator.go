package main

// https://dev.to/koddr/the-easiest-way-to-embed-static-files-into-a-binary-file-in-your-golang-app-no-external-dependencies-43pc

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	blobFileName string = "./box/blob.go"
	embedFolder  string = "./static"
)

// Define vars for build template
var conv = map[string]interface{}{"conv": fmtByteSlice}
var tmpl = template.Must(template.New("").Funcs(conv).Parse(`package box

// Code generated by go generate; DO NOT EDIT.

func init() {
    {{- range $name, $file := . }}
        box.Add("{{ $name }}", []byte{ {{ conv $file }} })
    {{- end }}
}`),
)

func fmtByteSlice(s []byte) string {
	builder := strings.Builder{}

	for _, v := range s {
		builder.WriteString(fmt.Sprintf("%d,", int(v)))
	}

	return builder.String()
}

func main() {
	// Checking directory with files
	if _, err := os.Stat(embedFolder); os.IsNotExist(err) {
		log.Fatal("Configs directory does not exists!")
	}

	// Create map for filenames
	configs := make(map[string][]byte)

	// Walking through embed directory
	err := filepath.Walk(embedFolder, func(path string, info os.FileInfo, err error) error {
		relativePath := filepath.ToSlash(strings.TrimPrefix(path, embedFolder))

		if info.IsDir() {
			// Skip directories
			log.Println(path, "is a directory, skipping...")
			return nil
		} else {
			// If element is a simple file, embed
			log.Println(path, "is a file, packing in...")

			b, err := ioutil.ReadFile(path)
			if err != nil {
				// If file not reading
				log.Printf("Error reading %s: %s", path, err)
				return err
			}

			// Add file name to map
			configs[relativePath] = b
		}

		return nil
	})
	if err != nil {
		log.Fatal("Error walking through embed directory:", err)
	}

	// Create blob file
	f, err := os.Create(blobFileName)
	if err != nil {
		log.Fatal("Error creating blob file:", err)
	}
	defer f.Close()

	// Create buffer
	builder := &bytes.Buffer{}

	// Execute template
	if err = tmpl.Execute(builder, configs); err != nil {
		log.Fatal("Error executing template", err)
	}

	// Formatting generated code
	data, err := format.Source(builder.Bytes())
	if err != nil {
		log.Fatal("Error formatting generated code", err)
	}

	// Writing blob file
	if err = ioutil.WriteFile(blobFileName, data, os.ModePerm); err != nil {
		log.Fatal("Error writing blob file", err)
	}
}
