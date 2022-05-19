package test

import (
	"testing"

	. "github.com/hoegaarden/pkg/testing"
)

func TestPackage(t *testing.T) {
	t.Parallel()

	NewJQer(t, Ytt{}).
		IsTrue(`false`)
}
