package testinfra

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// helper function to transfer files from local device to temporary packer instance
func (provisioner *Provisioner) uploadFiles(comm packer.Communicator, files []string, destDir string) error {
	// iterate through files to transfer
	for _, file := range files {
		// validate file existence
		if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
			log.Printf("The file does not exist at: %s, and will not be transferred", file)
			continue
		}

		// determine file content (io.Reader) from file path (string)
		fileBytes, err := os.ReadFile(file)
		if err != nil {
			log.Printf("the file at %s could not be read", file)
			return err
		}
		fileIo := bytes.NewReader(fileBytes)

		// upload file to destination dir
		destination := fmt.Sprintf("%s/%s", destDir, file)
		if err = comm.Upload(destination, fileIo, nil); err != nil {
			log.Printf("The file at %s could not be transferred to %s on the temporary Packer instance", file, destination)
			return err
		}
	}

	return nil
}
