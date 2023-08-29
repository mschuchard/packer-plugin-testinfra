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
	var err error

	// iterate through files to transfer
	for _, file := range files {
		// validate file existence
		if _, err = os.Stat(file); errors.Is(err, os.ErrNotExist) {
			log.Printf("the file does not exist at path: %s, and will not be transferred", file)
			continue
		}

		// determine file content (io.Reader) from file path (string)
		var fileBytes []byte
		if fileBytes, err = os.ReadFile(file); err != nil {
			log.Printf("the file at path '%s' could not be read, and will not be transferred", file)
			continue
		}
		fileIo := bytes.NewReader(fileBytes)

		// upload file to destination dir
		destination := fmt.Sprintf("%s/%s", destDir, file)
		if err = comm.Upload(destination, fileIo, nil); err != nil {
			log.Printf("the file at %s could not be transferred to %s on the temporary Packer instance", file, destination)
			continue
		}
	}

	// return most recent error
	// the logger displays the relevant debugging information, and this return is useful only in a nil comparable context, and not for specific error types
	return err
}
