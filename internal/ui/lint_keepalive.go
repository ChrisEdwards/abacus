package ui

// lintKeepAlive references helpers that are only exercised via tests so that
// unused linters stay quiet.
var (
	_ = preloadAllComments
	_ backgroundCommentLoadCompleteMsg
)
