package semanticrelease

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nightapes/go-semantic-release/pkg/config"
)

func TestSemanticRelease_WriteChangeLog(t *testing.T) {

	type args struct {
		changelogContent     string
		file                 string
		overwrite            bool
		maxChangelogFileSize int64
	}
	tests := []struct {
		config  *config.ReleaseConfig
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "MoveExisting",
			args: args{
				changelogContent:     "go-semantic-release-rocks!",
				file:                 "test1.changelog.md",
				overwrite:            false,
				maxChangelogFileSize: 0,
			},
			wantErr: false,
		},
		{
			name: "ValidWithOverwrite",
			args: args{
				changelogContent:     "go-semantic-release-rocks!",
				file:                 "test2.changelog.md",
				overwrite:            true,
				maxChangelogFileSize: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := os.Create(tt.args.file)
			if err != nil {
				t.Error(err)
			}

			releaser := &SemanticRelease{}
			if err := releaser.WriteChangeLog(tt.args.changelogContent, tt.args.file, tt.args.overwrite, tt.args.maxChangelogFileSize); (err != nil) != tt.wantErr {
				t.Errorf("WriteChangeLog() error = %v, wantErr %v", err, tt.wantErr)
			}

			name := strings.Join(strings.Split(tt.args.file, ".")[:len(strings.Split(tt.args.file, "."))-1], ".")

			files, err := filepath.Glob("./" + name + "*")
			if err != nil {
				t.Error(err)
			}

			if !tt.wantErr && !tt.args.overwrite && tt.args.maxChangelogFileSize == 0 && len(files) <= 1 {
				t.Errorf("WriteChangelog() = should create a copy of the existing changelog file")
			}

			if !tt.wantErr && tt.args.overwrite && len(files) > 1 {
				t.Errorf("WriteChangelog() = should not create a copy of the changelog file")
			}

			for _, i := range files {
				err := os.Remove(i)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}
