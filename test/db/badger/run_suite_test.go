package pgtest

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestBadgerTestSuite(t *testing.T) {
	suite.Run(t, new(BadgerDbTestSuite))
}
