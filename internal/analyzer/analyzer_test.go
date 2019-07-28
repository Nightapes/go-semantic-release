package analyzer_test

import (
	"testing"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzer(t *testing.T) {

	_, err := analyzer.New("unknown", config.ChangelogConfig{})
	assert.Error(t, err)

}
