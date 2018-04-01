package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
)

var (
	glmode = false
)

func init() {
	flag.BoolVar(&glmode, "gl", glmode, "Enable local GL display")
	flag.Parse()
}

type Viewer interface {
	Update(s string)
}

var shaderFile = "fragment.glsl"

// Monitor modification of the fragment shader and notify all viewers when it
// is updated
func monitorSource(viewers []Viewer) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("fsnotify:", err)
	}
	defer watcher.Close()
	if err := watcher.Add("."); err != nil {
		log.Fatalln("fsnotify add:", err)
	}
	current := ""
	check := true
	for {
		if check {
			bs, err := ioutil.ReadFile(shaderFile)
			source := string(bs)
			if err != nil {
				log.Println(shaderFile, "could not be read:", err)
				continue
			}
			if current != source && source != "" {
				current = source
				log.Println("Source updated")
				for _, v := range viewers {
					v.Update(source)
				}
			}
		}

		check = false
		select {
		case e := <-watcher.Events:
			fn := path.Clean(e.Name)
			if fn == shaderFile && (e.Op&(fsnotify.Create|fsnotify.Write)) != 0 {
				check = true
			}
		case err := <-watcher.Errors:
			log.Println("fsnotify watcher:", err)
		}
	}
}

func main() {
	// Create the shader file if it's missing, so we have something to render
	if _, err := os.Lstat(shaderFile); err != nil && os.IsNotExist(err) {
		write(shaderFile, defaultFragmentShader)
	} else if err != nil {
		log.Fatalln(shaderFile, "lstat error:", err)
	}

	var viewers []Viewer
	viewers = append(viewers, NewWebView())
	if glmode {
		viewers = append(viewers, NewGlView())
	}
	monitorSource(viewers)
}

func write(fname, data string) {
	fp, err := os.Create(fname)
	if err != nil {
		log.Fatalln(fname, err)
	}
	defer fp.Close()
	fp.Write([]byte(data))
}
