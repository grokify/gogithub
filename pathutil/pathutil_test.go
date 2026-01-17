package pathutil

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{"empty path", "", nil},
		{"simple file", "file.txt", nil},
		{"nested path", "dir/subdir/file.txt", nil},
		{"leading slash", "/file.txt", nil},
		{"traversal dotdot", "../file.txt", ErrPathTraversal},
		{"nested traversal", "dir/../file.txt", ErrPathTraversal},
		{"double traversal", "../../etc/passwd", ErrPathTraversal},
		{"hidden traversal", "dir/sub/../../../etc", ErrPathTraversal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.path)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate(%q) = %v, want %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"empty", "", ""},
		{"dot", ".", ""},
		{"slash", "/", ""},
		{"simple file", "file.txt", "file.txt"},
		{"leading slash", "/file.txt", "file.txt"},
		{"trailing slash", "dir/", "dir"},
		{"double slash", "dir//file.txt", "dir/file.txt"},
		{"nested", "dir/subdir/file.txt", "dir/subdir/file.txt"},
		{"leading nested", "/dir/subdir/file.txt", "dir/subdir/file.txt"},
		{"backslash", "dir\\file.txt", "dir/file.txt"},
		{"mixed slashes", "dir\\subdir/file.txt", "dir/subdir/file.txt"},
		{"redundant dots", "dir/./file.txt", "dir/file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.path)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidateAndNormalize(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr error
	}{
		{"valid simple", "file.txt", "file.txt", nil},
		{"valid leading slash", "/dir/file.txt", "dir/file.txt", nil},
		{"invalid traversal", "../file.txt", "", ErrPathTraversal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndNormalize(tt.path)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateAndNormalize(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateAndNormalize(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name  string
		elems []string
		want  string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"file.txt"}, "file.txt"},
		{"two elements", []string{"dir", "file.txt"}, "dir/file.txt"},
		{"three elements", []string{"dir", "subdir", "file.txt"}, "dir/subdir/file.txt"},
		{"with empty", []string{"dir", "", "file.txt"}, "dir/file.txt"},
		{"all empty", []string{"", "", ""}, ""},
		{"leading slash ignored", []string{"/dir", "file.txt"}, "dir/file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Join(tt.elems...)
			if got != tt.want {
				t.Errorf("Join(%v) = %q, want %q", tt.elems, got, tt.want)
			}
		})
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantDir  string
		wantFile string
	}{
		{"empty", "", "", ""},
		{"simple file", "file.txt", "", "file.txt"},
		{"nested", "dir/file.txt", "dir/", "file.txt"},
		{"deeply nested", "dir/subdir/file.txt", "dir/subdir/", "file.txt"},
		{"leading slash", "/dir/file.txt", "dir/", "file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDir, gotFile := Split(tt.path)
			if gotDir != tt.wantDir || gotFile != tt.wantFile {
				t.Errorf("Split(%q) = (%q, %q), want (%q, %q)", tt.path, gotDir, gotFile, tt.wantDir, tt.wantFile)
			}
		})
	}
}

func TestDir(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"empty", "", ""},
		{"simple file", "file.txt", ""},
		{"nested", "dir/file.txt", "dir"},
		{"deeply nested", "dir/subdir/file.txt", "dir/subdir"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Dir(tt.path)
			if got != tt.want {
				t.Errorf("Dir(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestBase(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"empty", "", ""},
		{"simple file", "file.txt", "file.txt"},
		{"nested", "dir/file.txt", "file.txt"},
		{"deeply nested", "dir/subdir/file.txt", "file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Base(tt.path)
			if got != tt.want {
				t.Errorf("Base(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestExt(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"no extension", "file", ""},
		{"simple", "file.txt", ".txt"},
		{"double extension", "file.tar.gz", ".gz"},
		{"hidden file", ".gitignore", ".gitignore"},
		{"nested", "dir/file.txt", ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Ext(tt.path)
			if got != tt.want {
				t.Errorf("Ext(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix string
		want   bool
	}{
		{"empty prefix", "dir/file.txt", "", true},
		{"exact match", "dir", "dir", true},
		{"nested match", "dir/file.txt", "dir", true},
		{"no match", "dir/file.txt", "other", false},
		{"partial no match", "directory/file.txt", "dir", false},
		{"deeply nested match", "dir/subdir/file.txt", "dir/subdir", true},
		{"prefix longer", "dir", "dir/subdir", false},
		{"both normalized", "/dir/file.txt", "/dir", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPrefix(tt.path, tt.prefix)
			if got != tt.want {
				t.Errorf("HasPrefix(%q, %q) = %v, want %v", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}
