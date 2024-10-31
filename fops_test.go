package sockfile

import (
	"os"
	"testing"
)

var (
	fo         *FileOps
	currentDir string
)

func init() {
	fo = new("./testData")
	currentDir, _ = os.Getwd()
}

func TestCompressZip(t *testing.T) {
	target := currentDir + "/testData/zipTest.zip"
	fo.removeIfExist(target)
	err := fo.compressZip(currentDir+"/testData/zipTest", target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnzip(t *testing.T) {
	target := currentDir + "/testData/zipTest"
	fo.removeIfExist(target)
	err := fo.unzip(target + ".zip")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTargz(t *testing.T) {
	target := currentDir + "/testData/tarTest.tar.gz"
	fo.removeIfExist(target)
	err := fo.targz(currentDir+"/testData/tarTest", target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnTargz(t *testing.T) {
	target := currentDir + "/testData/tarTest"
	fo.removeIfExist(target)

	err := fo.untargz(target + ".tar.gz")
	if err != nil {
		t.Fatal(err)
	}
}
