package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdaashir/NSM/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeWrite(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Test normal write
	data := []byte("test data")
	err := utils.SafeWrite(testFile, data, 0600)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, data, content)

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			err := utils.SafeWrite(testFile, []byte("test"), 0600)
			if err != nil {
				t.Errorf("SafeWrite failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all writes
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestFileLock(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "lock-test.txt")

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			lock := utils.AcquireLock(tmpFile)
			defer lock.Release()
			// Simulate work
			if err := utils.SafeWrite(tmpFile, []byte("test"), 0600); err != nil {
				t.Errorf("SafeWrite failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid relative", "foo/bar.txt", false},
		{"directory traversal", "../foo.txt", true},
		{"absolute path", "/etc/passwd", true},
		{"current directory", ".", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidatePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")

	// Create source file
	testData := []byte("test data")
	require.NoError(t, utils.SafeWrite(srcFile, testData, 0644))

	// Test copy
	err := utils.CopyFile(srcFile, dstFile)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, testData, content)

	// Verify permissions
	srcInfo, err := os.Stat(srcFile)
	require.NoError(t, err)
	dstInfo, err := os.Stat(dstFile)
	require.NoError(t, err)
	assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
}

func TestIsEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Test empty directory
	empty, err := utils.IsEmptyDir(tmpDir)
	require.NoError(t, err)
	assert.True(t, empty)

	// Add file and test non-empty
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, utils.SafeWrite(testFile, []byte("test"), 0644))

	empty, err = utils.IsEmptyDir(tmpDir)
	require.NoError(t, err)
	assert.False(t, empty)
}
