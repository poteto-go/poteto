package poteto

import "testing"

func BenchmarkInsertAndSearch(b *testing.B) {
	urls := []string{
		"/",
		"/example.com/v1/users/find/poteto",
		"/example.com/v1/users/find/potato",
		"/example.com/v1/users/find/jagaimo",
		"/example.com/v1/users/create/poteto",
		"/example.com/v1/users/create/potato",
		"/example.com/v1/users/create/jagaimo",
		"/example.com/v1/members/find/poteto",
		"/example.com/v1/members/find/potato",
		"/example.com/v1/members/find/jagaimo",
		"/example.com/v1/members/create/poteto",
		"/example.com/v1/members/create/potato",
		"/example.com/v1/members/create/jagaimo",
		"/example.com/v2/users/find/poteto",
		"/example.com/v2/users/find/potato",
		"/example.com/v2/users/find/jagaimo",
		"/example.com/v2/users/create/poteto",
		"/example.com/v2/users/create/potato",
		"/example.com/v2/users/create/jagaimo",
		"/example.com/v2/members/find/poteto",
		"/example.com/v2/members/find/potato",
		"/example.com/v2/members/find/jagaimo",
		"/example.com/v2/members/create/poteto",
		"/example.com/v2/members/create/potato",
		"/example.com/v2/members/create/jagaimo",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Insert
		route := NewRoute().(*route)
		for _, url := range urls {
			route.Insert(url, nil)
		}

		// Search
		for _, url := range urls {
			route.Search(url)
		}
	}
}
