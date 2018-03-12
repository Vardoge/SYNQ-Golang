package common

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CopyDir will walk through each file in srcDir, and call
// CopyFile on that file into dstDir.
func CopyDir(srcDir string, dstDir string) (err error) {
	cropNDirs := len(strings.Split(srcDir, "/"))
	walkFn := func(iFilepath string, info os.FileInfo, err error) error {

		// swap out the directory 'content_sanctuary' with 'content'
		inDirs := strings.Split(iFilepath, "/")
		outDirs := append([]string{dstDir}, inDirs[cropNDirs:]...)

		dstFilepath := path.Join(outDirs...)

		if info.Mode().IsDir() {
			os.MkdirAll(dstFilepath, 0755)
		} else {

			if !info.Mode().IsRegular() {
				return nil
			}

			if err != nil {
				return err
			}

			err = CopyFile(iFilepath, dstFilepath)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = filepath.Walk(srcDir, walkFn)
	if err != nil {
		return err
	}
	return nil
}

// Copied and modified from https://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Copy the file contents from src to dst.
func CopyFile(src string, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}

	}
	if os.SameFile(sfi, dfi) {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
