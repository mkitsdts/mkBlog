package static

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed avatar.jpg
var defaultAvatar []byte

// EnsureDefaultAvatar writes the embedded avatar into the data directory when
// the user has not provided one yet.
func EnsureDefaultAvatar(dataPath string) (err error) {
	avatarPath := filepath.Join(dataPath, "avatar.jpg")
	file, err := os.OpenFile(avatarPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create %s: %w", avatarPath, err)
	}

	written := false
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close %s: %w", avatarPath, closeErr)
		}
		if !written || err != nil {
			_ = os.Remove(avatarPath)
		}
	}()

	if _, err = file.Write(defaultAvatar); err != nil {
		return fmt.Errorf("write %s: %w", avatarPath, err)
	}
	written = true
	return nil
}
