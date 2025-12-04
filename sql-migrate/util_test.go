package migrate

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsSupportFilename(t *testing.T) {
	require.False(t, IsSupportFilename("V__init_sql.sql"))
	require.True(t, IsSupportFilename("V1__init_sql.sql"))
}

func TestSplitFilename(t *testing.T) {
	require.Equal(t, "12345", SplitFilename("V12345__init_sql.sql"))
}
