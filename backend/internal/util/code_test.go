package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAliasValidation(t *testing.T) {
	require.True(t, IsValidAlias("my-blog_2"))
	require.True(t, IsValidShortCode("abc1234"))
	require.True(t, IsValidShortCode("my-blog"))
	require.False(t, IsValidAlias("ab"))
	require.False(t, IsValidAlias("-blog"))
	require.False(t, IsValidAlias("my blog"))
	require.False(t, IsValidAlias("this-alias-is-too-long"))
}

func TestNormalizeTags(t *testing.T) {
	tags, err := NormalizeTags([]string{" Launch ", "launch", "email_campaign"})
	require.NoError(t, err)
	require.Equal(t, []string{"launch", "email_campaign"}, tags)

	_, err = NormalizeTags([]string{"invalid tag"})
	require.Error(t, err)
}
