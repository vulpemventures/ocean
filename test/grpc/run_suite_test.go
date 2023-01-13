package pgtest

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestGrpcTestSuite(t *testing.T) {
	suite.Run(t, new(GrpcDbTestSuite))
}
