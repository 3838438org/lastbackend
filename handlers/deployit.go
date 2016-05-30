package handlers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/deployithq/deployit/env"
	"github.com/deployithq/deployit/utils"
	"gopkg.in/urfave/cli.v2"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Main deploy it handler
// - Archive all files which is in folder
// - Send it to server

func DeployIt(c *cli.Context) error {

	env := NewEnv()

	var archiveName string = "tar.gz"
	var archivePath string = fmt.Sprintf("%s/.dit/%s", env.Path, archiveName)

	splittedPath := strings.Split(env.Path, "/")

	appName := splittedPath[len(splittedPath)-1]

	env.Log.Infof("Creating app: %s", appName)

	// Creating archive
	fw, err := os.Create(archivePath)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	gw := gzip.NewWriter(fw)
	tw := tar.NewWriter(gw)

	// Deleting archive after function ends
	defer func() {
		env.Log.Debug("Deleting archive: ", archivePath)

		fw.Close()
		gw.Close()
		tw.Close()

		// Deleting files
		err = os.Remove(archivePath)
		if err != nil {
			env.Log.Error(err)
			return
		}
	}()

	// Listing all files from database to know what files were deleted from previous run
	storedFiles, err := env.Storage.ListAllFiles(env.Log)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	// TODO Include deleted folders to deletedFiles like "nginx/"

	env.Log.Info("Packing files")
	storedFiles, err = PackFiles(env, tw, env.Path, storedFiles)
	if err != nil {
		return err
	}

	deletedFiles := []string{}

	for key, _ := range storedFiles {
		env.Log.Debug("Deleting: ", key)
		err = env.Storage.Delete(env.Log, key)
		if err != nil {
			return err
		}
		deletedFiles = append(deletedFiles, key)
	}

	tw.Close()
	gw.Close()
	fw.Close()

	bodyBuffer := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuffer)

	// Adding deleted files to request
	if len(deletedFiles) > 0 {
		delFiles, err := json.Marshal(deletedFiles)
		if err != nil {
			env.Log.Error(err)
			return err
		}

		bodyWriter.WriteField("deleted", string(delFiles))
	}

	// Adding application info to request
	bodyWriter.WriteField("name", appName)
	bodyWriter.WriteField("tag", Tag)

	archiveInfo, err := os.Stat(archivePath)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	// If archive size is 32 it means that it is empty and we don't need to send it
	if archiveInfo.Size() != 32 {
		fh, err := os.Open(archivePath)
		if err != nil {
			env.Log.Error(err)
			return err
		}

		fileWriter, err := bodyWriter.CreateFormFile("file", "tar.gz")
		if err != nil {
			env.Log.Error(err)
			return err
		}

		_, err = io.Copy(fileWriter, fh)
		if err != nil {
			env.Log.Error(err)
			return err
		}

		fh.Close()
	}

	bodyWriter.Close()

	env.Log.Debugf("%s/app/deploy", Host)

	// Creating response for file uploading with fields
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/app/deploy", Host), bodyBuffer)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	env.Log.Infof("Uploading sources")

	// TODO Show uploading progress

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	// TODO Stream build

	// Reading response from server
	resp_body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	// TODO Handle errors from http - clear DB if was first run

	env.Log.Debug(res.Status)
	env.Log.Debug(string(resp_body))

	res.Body.Close()

	env.Log.Infof("Done")

	return nil
}

func PackFiles(env *env.Env, tw *tar.Writer, filesPath string, storedFiles map[string]string) (map[string]string, error) {

	// Opening directory with files
	dir, err := os.Open(filesPath)

	if err != nil {
		env.Log.Error(err)
		return storedFiles, err
	}

	// Reading all files
	files, err := dir.Readdir(0)
	if err != nil {
		env.Log.Error(err)
		return storedFiles, err
	}

	for _, file := range files {

		fileName := file.Name()

		currentFilePath := fmt.Sprintf("%s/%s", filesPath, fileName)

		// Ignoring files which is not needed for build to make archive smaller
		// TODO: create base .ignore file on first application creation
		// TODO Create exclude lib
		// TODO Parse .gitignore and exclude files from
		if fileName == ".git" || fileName == ".idea" || fileName == ".dit" || fileName == "node_modules" {
			continue
		}

		// If it was directory - calling this function again
		// In other case adding file to archive
		if file.IsDir() {
			storedFiles, err = PackFiles(env, tw, currentFilePath, storedFiles)
			if err != nil {
				return storedFiles, err
			}
			continue
		}

		// Creating path, which will be inside of archive
		newPath := strings.Replace(currentFilePath, env.Path, "", 1)[1:]

		// Creating hash
		hash := utils.Hash(fmt.Sprintf("%s:%s:%s", file.Name(), strconv.FormatInt(file.Size(), 10), file.ModTime()))

		delete(storedFiles, newPath)

		if storedFiles[newPath] == hash {
			continue
		}

		// If hashes are not equal - add file to archive
		env.Log.Debug("Packing file: ", currentFilePath)

		err = env.Storage.Write(env.Log, newPath, hash)
		if err != nil {
			return storedFiles, err
		}

		fr, err := os.Open(currentFilePath)
		if err != nil {
			env.Log.Error(err)
			return storedFiles, err
		}

		h := &tar.Header{
			Name:    newPath,
			Size:    file.Size(),
			Mode:    int64(file.Mode()),
			ModTime: file.ModTime(),
		}

		err = tw.WriteHeader(h)
		if err != nil {
			env.Log.Error(err)
			return storedFiles, err
		}

		_, err = io.Copy(tw, fr)
		if err != nil {
			env.Log.Error(err)
			return storedFiles, err
		}

		fr.Close()

	}

	dir.Close()

	return storedFiles, err

}
