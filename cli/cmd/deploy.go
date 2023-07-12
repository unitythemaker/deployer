package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the build output to the Bulut server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return checkLogin()
	},
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
func createUploadRequest(filePath, url, suffix, apiKey string) (*http.Request, error) {
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

	req, err := http.NewRequest("PUT", url+suffix, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", apiKey)

	return req, nil
}

// TODO: Move to a common place
func getApiKeyForServer(server string) (string, error) {
	apiKey, err := ring.Get(fmt.Sprintf("api-key_%s", server))
	if err != nil {
		return "", err
	}
	return apiKey, nil
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

func checkDeployment(namespaceName, deploymentName string) (bool, error) {
	serverURL := getServerURL()
	deploymentURL := fmt.Sprintf("%s/deployment/%s/%s", serverURL, namespaceName, deploymentName)
	apiKey, err := getApiKeyForServer(serverURL)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("GET", deploymentURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		errResp := &bytes.Buffer{}
		_, err := io.Copy(errResp, resp.Body)
		if err != nil {
			return false, err
		}

		fmt.Printf("bad status: %s, response: %s\n", resp.Status, errResp)
		return false, err
	}

	return true, nil
}

// Server-side type
//type CreateDeploymentRequest struct {
//	Name      string `json:"name" form:"name"`
//	Namespace string `json:"namespace" form:"namespace"`
//}

func createDeployment(namespaceName, deploymentName string) error {
	serverURL := getServerURL()
	deploymentURL := fmt.Sprintf("%s/deployment/", serverURL)
	apiKey, err := getApiKeyForServer(serverURL)
	if err != nil {
		return err
	}

	formData := url.Values{}
	formData.Add("name", deploymentName)
	formData.Add("namespace", namespaceName)

	req, err := http.NewRequest("POST", deploymentURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errResp := &bytes.Buffer{}
		_, err := io.Copy(errResp, resp.Body)
		if err != nil {
			return err
		}

		fmt.Printf("bad status: %s, response: %s\n", resp.Status, errResp)
		return err
	}

	return nil
}

func deploy(args []string) error {
	serverURL := getServerURL()
	buildPath := viper.GetString("config.build-path")
	entrypoint := viper.GetString("config.entrypoint")
	deploymentName := viper.GetString("deployment.name")
	namespace := viper.GetString("deployment.namespace")

	if namespace == "" {
		prompt := &survey.Input{
			Message: "New deployment namespace:",
		}
		err := survey.AskOne(prompt, &namespace, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}
	if deploymentName == "" {
		prompt := &survey.Input{
			Message: "New deployment name:",
		}
		err := survey.AskOne(prompt, &deploymentName, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}

	// Check if deployment already exists
	deploymentExists, err := checkDeployment(namespace, deploymentName)
	if err != nil {
		return err
	}
	if deploymentExists {
		fmt.Printf("Deployment %s/%s already exists.\n", namespace, deploymentName)
		fmt.Println("Continuing to update deployment")
	} else {
		fmt.Printf("Creating new deployment %s/%s\n", namespace, deploymentName)
		err = createDeployment(namespace, deploymentName)
		if err != nil {
			return err
		}
		fmt.Println("Deployment created")
	}

	// Create zip file from build output
	zipFilename := generateTempFilename()

	if err := zipDirectory(buildPath, zipFilename); err != nil {
		return err
	}

	// Get API key
	apiKey, err := getApiKeyForServer(serverURL)
	if err != nil {
		return err
	}

	// Prepare upload request
	urlSuffix := fmt.Sprintf("/deployment/upload/%s/%s", namespace, deploymentName)
	uploadRequest, err := createUploadRequest(zipFilename, serverURL, urlSuffix, apiKey)

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
