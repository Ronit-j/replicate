package storage

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type DiskStorage struct {
	folder string
}

func NewDiskStorage(folder string) (*DiskStorage, error) {
	return &DiskStorage{
		folder: folder,
	}, nil
}

// Get data at path
func (s *DiskStorage) Get(p string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(s.folder, p))
}

// Put data at path
func (s *DiskStorage) Put(p string, data []byte) error {
	fullPath := path.Join(s.folder, p)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fullPath, data, 0644)
}

// PutDirectory recursively puts the local `localPath` directory into path `storagePath` on storage
//
// Parallels Storage.put_directory in Python.
func (s *DiskStorage) PutDirectory(localPath string, storagePath string) error {
	files, err := putDirectoryFiles(localPath, storagePath)
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(file.Source)
		if err != nil {
			return err
		}
		err = s.Put(file.Dest, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *DiskStorage) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	err := filepath.Walk(path.Join(s.folder, folder), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == filename {
			relPath, err := filepath.Rel(s.folder, path)
			if err != nil {
				return err
			}
			results <- ListResult{Path: relPath}
		}
		return nil
	})
	if err != nil {
		// If directory does not exist, treat this as empty. This is consistent with how blob storage
		// would behave
		if os.IsNotExist(err) {
			close(results)
			return
		}

		results <- ListResult{Error: err}
	}
	close(results)
}
