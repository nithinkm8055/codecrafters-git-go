package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type GitObjectHeader struct {
	objectType string
	size       int
}

func parseGitObject(content string) (*GitObjectHeader, string, error) {
	// Git object header format is: "<object_type> <size>\0"
	parts := strings.SplitN(content, " ", 2)
	if len(parts) < 2 {
		return nil, "", fmt.Errorf("invalid object header")
	}

	objectType := parts[0]
	rest := parts[1]

	// The rest of the string contains the size and the actual content, so we need to locate the null byte separator
	nullByteIndex := strings.IndexByte(rest, 0)
	if nullByteIndex == -1 {
		return nil, "", fmt.Errorf("invalid object format (missing null byte)")
	}

	sizeStr := rest[:nullByteIndex]
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid size in object header: %v", err)
	}

	// The object data starts after the null byte
	objectData := rest[nullByteIndex+1:]

	return &GitObjectHeader{objectType: objectType, size: size}, objectData, nil
}

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

				// TODO: revisit and simplify this
				content := ""
				flag := false
				for i := range readFileContent {
					if readFileContent[i] == 0 {
						flag = true
					}
					if flag && readFileContent[i] != 0 {
						content += string(readFileContent[i])
					}
				}

				fmt.Print(content)
			}
		}
	case "hash-object":
		if len(os.Args) > 2 {
			file, _ := os.ReadFile(os.Args[3])
			stats, _ := os.Stat(os.Args[3])
			content := string(file)
			contentAndHeader := fmt.Sprintf("blob %d\x00%s", stats.Size(), content)
			sha := (sha1.Sum([]byte(contentAndHeader)))
			hash := fmt.Sprintf("%x", sha)
			blobName := []rune(hash)
			blobPath := ".git/objects/"
			for i, v := range blobName {
				blobPath += string(v)
				if i == 1 {
					blobPath += "/"
				}
			}
			var buffer bytes.Buffer
			z := zlib.NewWriter(&buffer)
			z.Write([]byte(contentAndHeader))
			z.Close()
			os.MkdirAll(filepath.Dir(blobPath), os.ModePerm)
			f, _ := os.Create(blobPath)
			defer f.Close()
			f.Write(buffer.Bytes())
			fmt.Print(hash)
		}
	case "ls-tree":
		if len(os.Args) > 2 {
			// nameOnly := os.Args[2]
			hash := os.Args[3]

			dirName := hash[:2]
			fileName := hash[2:]

			if err := os.Chdir(fmt.Sprintf(".git/objects/%s", dirName)); err != nil {
				fmt.Fprintf(os.Stderr, "specified hash %s does not exist\n", hash)
				return
			}

			readFileContent, _ := DecompressAndRead(fileName)
			// Git object header processing (e.g., tree or blob)
			// A Git object is composed of a header and compressed content
			header, objectData, err := parseGitObject(content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing git object: %s\n", err)
				return
			}

			// Process the object based on its type
			if header.objectType == "tree" {
				// Handle tree object (directory)
				fmt.Println("This is a tree object")
				// Process the tree object (listing its contents, for example)
				fmt.Println(objectData)
			} else if header.objectType == "blob" {
				// Handle blob object (file content)
				fmt.Println("This is a blob object")
				fmt.Println(objectData)
			} else {
				fmt.Println("Unsupported object type")
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}

	// cd to .git/objects
	//
}
