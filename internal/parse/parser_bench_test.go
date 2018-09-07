package parse

import "testing"

func BenchmarkParse(t *testing.B) {
	for i := 0; i < t.N; i++ {
		ParseCommand("GET", "httpbin.org/post",
			"auth:token", "accept:application/json",
			"x-something:hi", "x-something-else:bye",
			"person.name=brett", "person.age:=100",
			"request-id=asdfadfadsfsdf")
	}
}
