package main

import (
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"
)

func DecompressAndRead(fileName string) (string, error) {
	compressedFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return "", errors.New("")
	}
	defer compressedFile.Close()

	// Create a zlib reader
	zlibReader, err := zlib.NewReader(compressedFile)
	if err != nil {
		fmt.Println("Error creating zlib reader:", err)
		return "", errors.New("")
	}
	defer zlibReader.Close()

	// Read decompressed data
	decompressedData, err := io.ReadAll(zlibReader)
	if err != nil {
		fmt.Println("Error reading decompressed data:", err)
		return "", errors.New("")
	}

	// Print the decompressed content
	return string(decompressedData), nil
}

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		// Uncomment this block to pass the first stage!
		//
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")
	case "cat-file":
		if len(os.Args) > 2 {
			pArg := os.Args[2]
			if pArg == "-p" {
				hash := os.Args[3]
				dirName := hash[:2]
				fileName := hash[2:]

				if err := os.Chdir(fmt.Sprintf(".git/objects/%s", dirName)); err != nil {
					fmt.Fprintf(os.Stderr, "specified hash %s does not exist\n", hash)
					return
				}

				readFileContent, err := DecompressAndRead(fileName)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err.Error())
					return
				}
				content := ""
				flag := false
				for i := range readFileContent {
					if readFileContent[i] == 0 {
						flag = true
					}
					if flag {
						content += string(readFileContent[i])
					}
				}

				fmt.Print(content)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}

	// cd to .git/objects
	//
}
