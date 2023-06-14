package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	deployCmd.PersistentFlags().String("build-path", ".output", "Filename of the build output file")
	deployCmd.PersistentFlags().String("entrypoint", "server/index.mjs", "Entrypoint of the build output file")

	viper.BindPFlag("build-path", deployCmd.Flags().Lookup("build-path"))
	viper.BindPFlag("entrypoint", deployCmd.Flags().Lookup("entrypoint"))
}

// TODO: Move to a common place
func generateTempFilename() string {
	rand.Seed(time.Now().UnixNano())
	randID := fmt.Sprintf("%016x", rand.Uint64())
	return filepath.Join(os.TempDir(), "deployment-"+randID)
}

// TODO: Move to a common place
func createUploadRequest(filePath, url string) (*http.Request, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))

	if err != nil {
		return nil, err
	}

	bar := pb.Full.Start64(fileInfo.Size())
	defer bar.Finish()

	barReader := bar.NewProxyReader(file)
	_, err = io.Copy(part, barReader)

	if err = writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

// TODO: Move to a common place
func uploadFile(request *http.Request) error {
	client := &http.Client{}
	resp, err := client.Do(request)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errResp := &bytes.Buffer{}
		_, err := io.Copy(errResp, resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("bad status: %s, response: %s", resp.Status, errResp)
	}
	fmt.Println("Upload successful. Deploy in progress!")
	return nil
}

// TODO: Move to a common place
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

		pathInZip, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		if len(pathInZip) < 1 {
			return nil
		}
		if pathInZip[0] != '/' {
			pathInZip = "/" + pathInZip
		}

		if info.IsDir() {
			pathInZip = fmt.Sprintf("%s%c", pathInZip, os.PathSeparator)
			_, err = zipWriter.Create(pathInZip)
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipFile, err := zipWriter.Create(pathInZip)
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

// TODO: Move to a common place
func getServerURL() string {
	host := viper.GetString("server.host")
	port := viper.GetUint("server.port")
	ssl := viper.GetBool("server.ssl")

	protocol := "http"
	if ssl {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

func deploy(args []string) error {
	serverURL := getServerURL()
	buildPath := viper.GetString("config.build-path")
	entrypoint := viper.GetString("config.entrypoint")
	// Create zip file from build output
	zipFilename := generateTempFilename()

	if err := zipDirectory(buildPath, zipFilename); err != nil {
		return err
	}

	// Prepare upload request
	uploadRequest, err := createUploadRequest(zipFilename, serverURL+"/upload")

	// Add additional info to request
	query := uploadRequest.URL.Query()
	query.Add("entrypoint", entrypoint)
	uploadRequest.URL.RawQuery = query.Encode()

	if err != nil {
		return err
	}

	// Upload zip file to server
	err = uploadFile(uploadRequest)

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
