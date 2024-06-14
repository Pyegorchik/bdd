package integrationstests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuiteUser))
}
