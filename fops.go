package sockfile

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FileOps struct {
	dir string
}

func new(dir string) *FileOps {
	return &FileOps{dir}
}

func (fo *FileOps) compressZip(srcDir string, dstZipFile string) error {
	zipFile, err := os.OpenFile(dstZipFile, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	archive := zip.NewWriter(zipFile)
	defer archive.Close()
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if path == srcDir {
			return nil
		}
		info, _ := d.Info()
		h, _ := zip.FileInfoHeader(info)
		h.Name = strings.TrimPrefix(path, srcDir+"/")
		if info.IsDir() {
			h.Name += "/"
		} else {
			h.Method = zip.Deflate
		}
		writer, _ := archive.CreateHeader(h)
		if !info.IsDir() {
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			io.Copy(writer, srcFile)
		}
		return nil
	})
}

func (fo *FileOps) unzip(zipFile string) error {
	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer archive.Close()
	fileName := filepath.Base(zipFile)
	fileDir := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	dir := fmt.Sprintf("%s/%s", filepath.Dir(zipFile), fileDir)
	for _, f := range archive.File {
		filePath := filepath.Join(dir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to make directory (%v)", err)
		}
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file (%v)", err)
		}
		defer dstFile.Close()
		fileInArchive, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip (%v)", err)
		}
		defer fileInArchive.Close()
		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return fmt.Errorf("failed to copy file in zip (%v)", err)
		}
	}
	return nil
}

func (fo *FileOps) removeIfExist(target string) error {
	_, err := os.Stat(target)
	if err == nil {
		return os.RemoveAll(target)
	}
	return nil
}

func (fo *FileOps) targz(srcDir string, dstFile string) error {
	targzFile, err := os.OpenFile(dstFile, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer targzFile.Close()

	gz := gzip.NewWriter(targzFile)
	defer gz.Close()

	tarFile := tar.NewWriter(gz)
	defer tarFile.Close()

	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if path == srcDir {
			return nil
		}
		info, _ := d.Info()
		h, _ := tar.FileInfoHeader(info, "")
		h.Name = strings.TrimPrefix(path, srcDir+"/")
		if info.IsDir() {
			h.Name += "/"
		}
		tarFile.WriteHeader(h)
		if !info.IsDir() {
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			io.Copy(tarFile, srcFile)
		}
		return nil
	})
}

func (fo *FileOps) untargz(tgFile string) error {
	file, err := os.Open(tgFile)
	if err != nil {
		return err
	}
	defer file.Close()
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()
	archive := tar.NewReader(gzReader)

	fileName := filepath.Base(tgFile)
	fileDir := strings.TrimSuffix(fileName, ".tar.gz")
	dir := fmt.Sprintf("%s/%s", filepath.Dir(tgFile), fileDir)
	for header, err := archive.Next(); err != io.EOF; header, err = archive.Next() {
		if err != nil {
			return err
		}

		filePath := filepath.Join(dir, header.Name)
		if header.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to make directory (%v)", err)
		}
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
		if err != nil {
			return fmt.Errorf("failed to create file (%v)", err)
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, archive); err != nil {
			return fmt.Errorf("failed to copy file in tar.gz (%v)", err)
		}
	}
	return nil
}
