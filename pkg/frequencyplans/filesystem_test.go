// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/frequencyplans"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type frequencyPlansFileSystem string

func createMockFileSystem() (frequencyPlansFileSystem, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}

	f, err := os.Create(filepath.Join(dir, frequencyplans.DefaultListFilename))
	if err != nil {
		return "", err
	}
	f.Close()

	return frequencyPlansFileSystem(dir), nil
}

func (fs frequencyPlansFileSystem) Destroy() error {
	return os.RemoveAll(string(fs))
}

func (fs frequencyPlansFileSystem) Dir() string {
	return string(fs)
}

func TestReadEmptyFrequencyPlans(t *testing.T) {
	a := assertions.New(t)

	fs, err := createMockFileSystem()
	a.So(err, should.BeNil)
	defer fs.Destroy()

	rootPathOption := frequencyplans.FileSystemRootPathOption(fs.Dir())
	store, err := frequencyplans.ReadFileSystemStore(rootPathOption)
	a.So(err, should.BeNil)

	ids := store.GetAllIDs()
	a.So(len(ids), should.Equal, 0)
}

func TestAbsoluteListFilepath(t *testing.T) {
	a := assertions.New(t)

	fs, err := createMockFileSystem()
	a.So(err, should.BeNil)
	defer fs.Destroy()

	rootPathOption := frequencyplans.FileSystemRootPathOption(fs.Dir())
	store, err := frequencyplans.ReadFileSystemStore(rootPathOption)
	a.So(err, should.BeNil)

	ids := store.GetAllIDs()
	a.So(len(ids), should.Equal, 0)
}

func TestReadInexistantStore(t *testing.T) {
	a := assertions.New(t)

	tempDir, err := ioutil.TempDir("", "")
	a.So(err, should.BeNil)

	rootPathOption := frequencyplans.FileSystemRootPathOption(tempDir)
	_, err = frequencyplans.ReadFileSystemStore(rootPathOption)
	a.So(err, should.NotBeNil)
}

func TestReadInvalidStore(t *testing.T) {
	a := assertions.New(t)

	fs, err := createMockFileSystem()
	a.So(err, should.BeNil)
	defer fs.Destroy()

	listFile := filepath.Join(fs.Dir(), frequencyplans.DefaultListFilename)
	err = ioutil.WriteFile(listFile, []byte(`    invalid: yaml`), 0666)
	a.So(err, should.BeNil)

	rootPathOption := frequencyplans.FileSystemRootPathOption(fs.Dir())
	_, err = frequencyplans.ReadFileSystemStore(rootPathOption)
	a.So(err, should.NotBeNil)
}

func TestReadStoreWithInexistantFP(t *testing.T) {
	a := assertions.New(t)

	fs, err := createMockFileSystem()
	a.So(err, should.BeNil)
	defer fs.Destroy()

	listFile := filepath.Join(fs.Dir(), frequencyplans.DefaultListFilename)
	err = ioutil.WriteFile(listFile, []byte(dummyFPList()), 0666)
	a.So(err, should.BeNil)

	rootPathOption := frequencyplans.FileSystemRootPathOption(fs.Dir())
	_, err = frequencyplans.ReadFileSystemStore(rootPathOption)
	a.So(err, should.NotBeNil)
}

func TestReadValidStore(t *testing.T) {
	a := assertions.New(t)

	fs, err := createMockFileSystem()
	a.So(err, should.BeNil)
	defer fs.Destroy()

	listFile := filepath.Join(fs.Dir(), frequencyplans.DefaultListFilename)
	err = ioutil.WriteFile(listFile, []byte(dummyFPList()), 0666)
	a.So(err, should.BeNil)

	frequencyPlan, err := dummyEUFP()
	a.So(err, should.BeNil)

	frequencyPlanFilename := filepath.Join(fs.Dir(), euFPFilename)
	err = ioutil.WriteFile(frequencyPlanFilename, []byte(frequencyPlan), 0666)
	a.So(err, should.BeNil)

	rootPathOption := frequencyplans.FileSystemRootPathOption(fs.Dir())
	store, err := frequencyplans.ReadFileSystemStore(rootPathOption)
	a.So(err, should.BeNil)

	ids := store.GetAllIDs()
	a.So(len(ids), should.Equal, 1)

	storedFrequencyPlan, err := store.GetByID(ids[0])
	a.So(err, should.BeNil)
	a.So(storedFrequencyPlan.BandID, should.Equal, string(band.EU_863_870))
}

func TestReadStoreWithInvalidFP(t *testing.T) {
	a := assertions.New(t)

	fs, err := createMockFileSystem()
	a.So(err, should.BeNil)
	defer fs.Destroy()

	listFile := filepath.Join(fs.Dir(), frequencyplans.DefaultListFilename)
	err = ioutil.WriteFile(listFile, []byte(dummyFPList()), 0666)
	a.So(err, should.BeNil)

	frequencyPlanFilename := filepath.Join(fs.Dir(), euFPFilename)
	err = ioutil.WriteFile(frequencyPlanFilename, []byte(`    invalid
yaml`), 0666)
	a.So(err, should.BeNil)

	rootPathOption := frequencyplans.FileSystemRootPathOption(fs.Dir())
	_, err = frequencyplans.ReadFileSystemStore(rootPathOption)
	a.So(err, should.NotBeNil)
}

func TestReadLocalStore(t *testing.T) {
	a := assertions.New(t)

	fs, err := createMockFileSystem()
	a.So(err, should.BeNil)
	defer fs.Destroy()

	err = os.Chdir(fs.Dir())
	a.So(err, should.BeNil)

	fpContent, err := dummyEUFP()
	a.So(err, should.BeNil)

	f, err := ioutil.TempFile("", "")
	a.So(err, should.BeNil)
	defer os.Remove(f.Name())
	_, err = f.Write([]byte(fpContent))
	a.So(err, should.BeNil)
	f.Close()

	frequencyPlanList := fmt.Sprintf(`- id: RANDOM_ID
  file: %s`, f.Name())

	err = ioutil.WriteFile(frequencyplans.DefaultListFilename, []byte(frequencyPlanList), 0666)
	a.So(err, should.BeNil)

	store, err := frequencyplans.ReadFileSystemStore()
	a.So(err, should.BeNil)
	ids := store.GetAllIDs()
	a.So(len(ids), should.Equal, 1)
}
