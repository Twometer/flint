package config

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"time"
)

const configUpdateTimeout = 250 * time.Millisecond

type Watcher struct {
	path            string
	fsWatcher       *fsnotify.Watcher
	CurrentConfig   Config
	OnConfigChanged func()
}

func WatchConfig(path string) (Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return Watcher{}, err
	}

	err = fsWatcher.Add(path)
	if err != nil {
		return Watcher{}, err
	}

	watcher := Watcher{
		fsWatcher: fsWatcher,
		path:      path,
	}
	err = watcher.reloadConfig()
	if err != nil {
		return Watcher{}, err
	}

	go watcher.processEvents()
	return watcher, err
}

func (watcher *Watcher) processEvents() {
	eventTimer := time.NewTimer(0)
	pendingEvent := false

	for {
		select {
		case event, ok := <-watcher.fsWatcher.Events:
			if !ok {
				return
			}
			if !event.Has(fsnotify.Write) {
				continue
			}
			pendingEvent = true
			eventTimer.Reset(configUpdateTimeout) // avoid double-updates

		case err, ok := <-watcher.fsWatcher.Errors:
			if !ok {
				return
			}
			log.Printf("config watcher error: %v\n", err)

		case <-eventTimer.C:
			if !pendingEvent {
				continue
			}
			pendingEvent = false
			watcher.handleConfigChanged()
		}
	}
}

func (watcher *Watcher) handleConfigChanged() {
	err := watcher.reloadConfig()
	if err != nil {
		log.Printf("error: failed to reload config: %v\n", err)
		return
	}

	if watcher.OnConfigChanged != nil {
		watcher.OnConfigChanged()
	}
	log.Println("Applying changed configuration")
}

func (watcher *Watcher) reloadConfig() error {
	config, err := loadConfig(watcher.path)
	if err != nil {
		return err
	}

	watcher.CurrentConfig = config
	return nil
}

func (watcher *Watcher) Close() {
	_ = watcher.fsWatcher.Close()
}
