package pgtest

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestInMemoryTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryDbTestSuite))
}
