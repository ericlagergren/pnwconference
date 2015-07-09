package reload

import (
	"fmt"
	"html/template"
	"path/filepath"
	"sync"

	"github.com/EricLagerg/pnwconference/cleanup"

	"github.com/golang/glog"
	"golang.org/x/exp/inotify"
)

const templateExt = ".gohtml"

var (
	TmplMap = make(map[string]TmplName)

	templatePath = filepath.Join("templates")
	base         = filepath.Join(templatePath, "Base"+templateExt)
	watcher      *inotify.Watcher

	// NOTE: This MUST match reload/enum.go!
	Tmpls = Templates{
		[]*template.Template{
			// SiteRouter
			template.Must(template.ParseFiles(filepath.Join(templatePath, "Index"+templateExt), base)),
			template.Must(template.ParseFiles(filepath.Join(templatePath, "About"+templateExt), base)),
			template.Must(template.ParseFiles(filepath.Join(templatePath, "Login"+templateExt), base)),
			template.Must(template.ParseFiles(filepath.Join(templatePath, "Create"+templateExt), base)),
			template.Must(template.ParseFiles(filepath.Join(templatePath, "Signup"+templateExt), base)),
			template.Must(template.ParseFiles(filepath.Join(templatePath, "ThankYou"+templateExt), base)),
			template.Must(template.ParseFiles(filepath.Join(templatePath, "ErrorPage"+templateExt), base)),
		},
		watcher,
		&sync.Mutex{},
	}
)

// Templates encapsulates our template files so we can load them into
// memory when the app is first run, but still be able to hot swap them.
type Templates struct {
	Templates []*template.Template

	*inotify.Watcher
	*sync.Mutex
}

func init() {
	// Generate the reverse of our enum so we can work backwards and reload
	// a file with the name instead of the enum.
	//
	// The issue arises because we've decided to use an array instead of a map
	// to describe our in-memory templates. While this is more efficient, it
	// causes issues with our hot reloading because the name of the file
	// given to us from inotify is the string representation of the file's
	// name, and we can't match that up with the enum on the fly (or generate
	// code that does that using //go: generate). So, we run an init func that
	// generates a map of the names to the enum so we can work backwards to
	// reload the file.
	for i := 0; i < len(_TmplName_index)-1; i++ {
		key := _TmplName_name[_TmplName_index[i]:_TmplName_index[i+1]]
		TmplMap[key] = TmplName(i)
	}

	// Set up our watcher for hot reloads of modified files.
	watcher, err := inotify.NewWatcher()
	if err != nil {
		glog.Fatalln(err)
	}

	err = watcher.Watch(templatePath)
	if err != nil {
		glog.Fatalln(err)
	}

	Tmpls.Watcher = watcher
	Tmpls.Watch()

	cleanup.Register("reload", watcher.Close) // Close watcher.
}

func (t *Templates) Watch() {
	go func() {
		for {
			select {
			case ev := <-t.Watcher.Event:

				// Check for modifications.
				if ev.Mask&inotify.IN_CREATE != 0 ||
					ev.Mask&inotify.IN_MODIFY != 0 {

					glog.V(2).Infof("File: %s Event: %s. Hot reloading.",
						ev.Name, ev.String())

					if err := t.reload(ev.Name); err != nil {
						glog.Fatalln(err)
					}
				}

			case err := <-t.Watcher.Error:
				glog.Errorln(err)
			}
		}
	}()
}

func (t *Templates) reload(name string) error {

	// panic: template: redefinition of template "base"
	if name == base {
		var err error

		// Reload all with the updated "base".
		for name, _ := range TmplMap {
			err = t.reload(filepath.Join(templatePath, name+templateExt))
			if err != nil {
				glog.Errorln(err)
			}
		}
		return err
	}

	// Suffix is ".gohtml"
	if len(name) >= len(templateExt) &&
		name[len(name)-len(templateExt):] == templateExt {

		// Gather what would be the key in our template map.
		// 'name' is in the format: "path/identifier.extension",
		// so trim the 'path/' and the '.extension' to get the
		// name (minus new extension) used inside of our map.
		// We add +1 to len(templatePath) because Go excludes
		// the trailing slash in filepath.Join, and if we were to
		// add the trailing slash we'd end up having two path slashes
		// inside 'name'. Then, find its numerical constant value from
		// inside our map.
		key := TmplMap[name[len(templatePath)+1:len(name)-len(templateExt)]]

		tmpl := template.Must(template.ParseFiles(name, base))

		t.Lock()
		t.Templates[key] = tmpl
		t.Unlock()

		return nil
	}

	return fmt.Errorf("Cannot hot reload template file %s\n", name)
}
