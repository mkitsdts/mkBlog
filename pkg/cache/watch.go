package cache

import (
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

func WatchCacheFiles(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("failed to create file watcher", "error", err)
		return
	}

	if err := watcher.Add(path); err != nil {
		slog.Error("failed to watch directory", "error", err)
		return
	}

	for {
		select {
		case event := <-watcher.Events:
			//有事件发生就触发事件处理
			EventHandle(event, watcher)
		case err := <-watcher.Errors:
			slog.Error("failed to watch directory", "error", err)
		}
	}
}

func EventHandle(event fsnotify.Event, watcher *fsnotify.Watcher) {
	BuildAssetCache(globalAssetCache.root)
}
