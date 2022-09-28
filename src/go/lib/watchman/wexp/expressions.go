package wexp

// Term base expression term object
type Term []interface{}

/*
Match

Examples:

	Match("*.txt", "basename")
	Match("dir/*.txt", "wholename")
	Match("src/**\/*.java", "wholename")

https://facebook.github.io/watchman/docs/expr/match.html
https://facebook.github.io/watchman/docs/expr/match.html#wildmatch
https://facebook.github.io/watchman/docs/expr/match.html#case-sensitivity
*/
func Match(opts ...interface{}) Term {
	return append(Term{"match"}, opts...)
}

func Not(opts ...interface{}) Term {
	return append(Term{"not"}, opts...)
}

func AllOf(opts ...interface{}) Term {
	return append(Term{"allof"}, opts...)
}

func Type(opts ...interface{}) Term {
	return append(Term{"type"}, opts...)
}
