package models

import "github.com/humpback/gounits/compress/tarlib"
import "github.com/humpback/gounits/httpx"
import "github.com/humpback/gounits/system"
import "github.com/humpback/gounits/utils"
import yaml "gopkg.in/yaml.v2"

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	//ErrProjectNameInvalid is exported
	ErrProjectNameInvalid = errors.New("project name invalid")
	//ErrProjectPackageFileInvalid is exported
	ErrProjectPackageFileInvalid = errors.New("project package file invalid")
)

//ProjectJSON is exported
type ProjectJSON struct {
	Name        string `json:"Name"`
	HashCode    string `json:"HashCode"`
	PackageFile string `json:"PackageFile"`
	Timestamp   int64  `json:"Timestamp"`
}

//ProjectData is exported
type ProjectData struct {
	*ProjectJSON
	ComposeFile  string `json:"ComposeFile"`
	ComposeBytes []byte `json:"ComposeBytes"`
}

func validateProjectName(projectName string) error {

	if strings.TrimSpace(projectName) == "" {
		return ErrProjectNameInvalid
	}
	return nil
}

func validateComposeData(composeData string) error {

	var err error
	buffer := []byte(composeData)
	if len(buffer) == 0 {
		return fmt.Errorf("compose data format invalid, data is empty")
	}
	template := map[string]interface{}{}
	if err = yaml.Unmarshal(buffer, &template); err != nil {
		return fmt.Errorf("compose data format invalid, %s", err)
	}
	return nil
}

func projectComposeFilePath(rootPath string, projectName string, hashCode string) string {

	projectDir := filepath.Dir(projectFilePath(rootPath, projectName))
	composeFilePath := filepath.Clean(projectDir + "/" + hashCode + "/docker-compose.yaml")
	return composeFilePath
}

func projectFilePath(rootPath string, projectName string) string {

	projectFilePath := filepath.Clean(rootPath + "/" + projectName + "/project.json")
	return projectFilePath
}

func projectHashCode(projectName string, packageFile string, timeStamp int64) string {

	projectStr := fmt.Sprintf("%s-%s-%d", projectName, packageFile, timeStamp)
	sum := sha256.Sum256([]byte(projectStr))
	return fmt.Sprintf("%x", sum)
}

func savePackageFile(rootPath string, projectName string, packageFile string, packageFileBytes []byte) (string, error) {

	timeStamp := time.Now().Unix()
	fps := strings.Split(packageFile, ".")
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s%s%d", projectName, fps[0], timeStamp)))
	fps[0] = fps[0] + "-" + fmt.Sprintf("%x", sum)[:6]
	packageFile = strings.Join(fps, ".")

	packageFilePath := filepath.Clean(rootPath + "/_uploads/" + projectName + "/" + packageFile)
	packageDir := filepath.Dir(packageFilePath)
	if err := system.MakeDirectory(packageDir); err != nil {
		return "", err
	}

	f, err := os.OpenFile(packageFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}

	defer f.Close()
	//io.Copy(f, reader)
	if _, err := f.Write(packageFileBytes); err != nil {
		return "", err
	}
	return packageFile, nil
}

func removePackageFile(rootPath string, projectName string, packageFile string, removeAll bool) error {

	if !removeAll {
		if _, err := url.ParseRequestURI(packageFile); err == nil {
			packageFile = filepath.Base(packageFile)
		}
	}

	packageFileDir := filepath.Clean(rootPath + "/_uploads/" + projectName)
	if !removeAll {
		packageFilePath := filepath.Clean(packageFileDir + "/" + packageFile)
		return os.Remove(packageFilePath)
	}
	return os.RemoveAll(packageFileDir)
}

func extractPackageFile(rootPath string, projectName string, packageFile string, hashCode string) error {

	packageFilePath := filepath.Clean(rootPath + "/_uploads/" + projectName + "/" + packageFile)
	if _, err := url.ParseRequestURI(packageFile); err == nil {
		err = httpx.GetFile(context.Background(), packageFilePath, packageFile, nil, nil)
		if err != nil {
			return fmt.Errorf("project %s package %s download failure.", projectName, packageFile)
		}
	}

	if !system.FileExist(packageFilePath) {
		return fmt.Errorf("project %s package %s not exists.", projectName, packageFile)
	}

	composeFilePath := projectComposeFilePath(rootPath, projectName, hashCode)
	composeFileDir := filepath.Dir(composeFilePath)
	err := tarlib.TarAutoDeCompress(packageFilePath, composeFileDir)
	if err != nil {
		return fmt.Errorf("project %s package %s extract failure.", projectName, packageFile)
	}
	return nil
}

//ComposeStorage is exported
//compose project local storage.
type ComposeStorage struct {
	sync.RWMutex
	rootPath   string
	filterDirs []string
}

//NewComposeStorage is exported
func NewComposeStorage(rootPath string) (*ComposeStorage, error) {

	var err error
	rootPath, err = filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	if err = system.MakeDirectory(rootPath + "/_uploads"); err != nil {
		return nil, err
	}

	return &ComposeStorage{
		rootPath:   rootPath,
		filterDirs: []string{"_uploads"},
	}, nil
}

//ProjectJSON is exported
func (storage *ComposeStorage) ProjectJSON(projectName string) (*ProjectJSON, error) {

	if err := validateProjectName(projectName); err != nil {
		return nil, err
	}

	storage.Lock()
	defer storage.Unlock()
	projectFilePath := projectFilePath(storage.rootPath, projectName)
	data, err := ioutil.ReadFile(projectFilePath)
	if err != nil {
		if os.IsNotExist(err) || strings.Index(err.Error(), "no such file or directory") > 0 {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	projectJSON := &ProjectJSON{}
	err = json.Unmarshal(data, projectJSON)
	if err != nil {
		return nil, err
	}
	return projectJSON, nil
}

//SetProjectJSON is exported
func (storage *ComposeStorage) SetProjectJSON(projectJSON *ProjectJSON) error {

	if err := validateProjectName(projectJSON.Name); err != nil {
		return err
	}

	storage.Lock()
	defer storage.Unlock()
	data, err := json.Marshal(projectJSON)
	if err != nil {
		return err
	}

	projectFilePath := projectFilePath(storage.rootPath, projectJSON.Name)
	projectDir := filepath.Dir(projectFilePath)
	if err = system.MakeDirectory(projectDir); err != nil {
		return err
	}
	return ioutil.WriteFile(projectFilePath, data, 0777)
}

//ProjectSpecs is exported
func (storage *ComposeStorage) ProjectSpecs() ([]*ProjectData, error) {

	fis, err := ioutil.ReadDir(storage.rootPath)
	if err != nil {
		return nil, err
	}

	projectDataArray := []*ProjectData{}
	for _, fi := range fis {
		if fi.IsDir() && !utils.Contains(fi.Name(), storage.filterDirs) {
			projectData, _ := storage.ProjectSpec(fi.Name())
			if projectData != nil {
				projectDataArray = append(projectDataArray, projectData)
			}
		}
	}
	return projectDataArray, nil
}

//ProjectSpec is exported
func (storage *ComposeStorage) ProjectSpec(projectName string) (*ProjectData, error) {

	projectJSON, err := storage.ProjectJSON(projectName)
	if err != nil {
		return nil, err
	}

	storage.Lock()
	defer storage.Unlock()
	composeFilePath := projectComposeFilePath(storage.rootPath, projectName, projectJSON.HashCode)
	composeBytes, err := ioutil.ReadFile(composeFilePath)
	if err != nil {
		return nil, err
	}

	err = validateComposeData(string(composeBytes))
	if err != nil {
		return nil, err
	}

	return &ProjectData{
		ProjectJSON:  projectJSON,
		ComposeFile:  composeFilePath,
		ComposeBytes: composeBytes,
	}, nil
}

//CreateProjectSpec is exported
func (storage *ComposeStorage) CreateProjectSpec(createProject CreateProject) (*ProjectData, error) {

	var (
		err         error
		projectJSON *ProjectJSON
	)

	if err = validateProjectName(createProject.Name); err != nil {
		return nil, err
	}

	if err = validateComposeData(createProject.ComposeData); err != nil {
		return nil, err
	}

	projectJSON, err = storage.ProjectJSON(createProject.Name)
	if projectJSON != nil || !os.IsNotExist(err) {
		return nil, os.ErrExist
	}

	timeStamp := time.Now().Unix()
	hashCode := projectHashCode(createProject.Name, createProject.PackageFile, timeStamp)
	projectJSON = &ProjectJSON{
		Name:        createProject.Name,
		HashCode:    hashCode,
		PackageFile: createProject.PackageFile,
		Timestamp:   timeStamp,
	}

	if err = storage.SetProjectJSON(projectJSON); err != nil {
		return nil, err
	}

	storage.Lock()
	defer func() {
		storage.Unlock()
		if err != nil {
			storage.RemoveProjectSpec(projectJSON.Name)
		}
	}()

	composeFilePath := projectComposeFilePath(storage.rootPath, projectJSON.Name, hashCode)
	composeFileDir := filepath.Dir(composeFilePath)
	err = system.MakeDirectory(composeFileDir)
	if err != nil {
		return nil, err
	}

	composeBytes := []byte(createProject.ComposeData)
	err = ioutil.WriteFile(composeFilePath, composeBytes, 0777)
	if err != nil {
		return nil, err
	}

	if projectJSON.PackageFile != "" {
		err = extractPackageFile(storage.rootPath, projectJSON.Name, projectJSON.PackageFile, projectJSON.HashCode)
		if err != nil {
			return nil, err
		}
	}

	return &ProjectData{
		ProjectJSON:  projectJSON,
		ComposeFile:  composeFilePath,
		ComposeBytes: composeBytes,
	}, nil
}

//RemoveProjectSpec is exported
func (storage *ComposeStorage) RemoveProjectSpec(projectName string) error {

	projectJSON, err := storage.ProjectJSON(projectName)
	if err != nil {
		return err
	}

	storage.Lock()
	defer storage.Unlock()
	if projectJSON.PackageFile != "" {
		removePackageFile(storage.rootPath, projectJSON.Name, projectJSON.PackageFile, false)
	}
	projectDir := filepath.Dir(projectFilePath(storage.rootPath, projectJSON.Name))
	return os.RemoveAll(projectDir)
}

//SaveProjectPackageFile is exported
func (storage *ComposeStorage) SaveProjectPackageFile(projectName string, packageFile string, packageFileBytes []byte) (string, error) {

	if err := validateProjectName(projectName); err != nil {
		return "", err
	}

	if packageFile == "" {
		return "", ErrProjectPackageFileInvalid
	}
	return savePackageFile(storage.rootPath, projectName, packageFile, packageFileBytes)
}
