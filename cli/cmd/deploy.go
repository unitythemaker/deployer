package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the build output to the Deployer server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deploy(args)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func generateTempFilename() string {
	rand.Seed(time.Now().UnixNano())
	randID := fmt.Sprintf("%016x", rand.Uint64())
	return filepath.Join(os.TempDir(), "deployment-"+randID)
}

func uploadFile(filePath, url string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	bar := pb.Full.Start64(fileInfo.Size())
	defer bar.Finish()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	barReader := bar.NewProxyReader(file)
	_, err = io.Copy(part, barReader)

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	fmt.Println("Upload successful. Deploy in progress!")
	return nil
}

func zipDirectory(dirPath, zipFilePath string) error {
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			path = fmt.Sprintf("%s%c", path, os.PathSeparator)
			_, err = zipWriter.Create(path)
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipFile, err := zipWriter.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func deploy(args []string) error {
	// Create zip file from build output
	zipFilename := generateTempFilename()
	err := zipDirectory(".output", zipFilename)
	if err != nil {
		return err
	}

	// Upload zip file to server
	err = uploadFile(zipFilename, serverURL+"/upload")
	if err != nil {
		return err
	}

	// Delete zip file
	err = os.Remove(zipFilename)
	if err != nil {
		return err
	}

	return nil
}
