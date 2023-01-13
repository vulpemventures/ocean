package pgtest

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestPgTestSuite(t *testing.T) {
	suite.Run(t, new(PgDbTestSuite))
}
