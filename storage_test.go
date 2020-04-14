package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github/luckylukas/whenthengo/types"
	"net/http"
	"strings"
	"testing"
)

func TestStorage_GetWhenThenKey(t *testing.T) {
	s := InMemoryStore{}
	assert.Equal(t, s.getWhenThenKey(&types.WhenThen{
		When: types.When{
			URL:    "/path/123",
			Method: "GET",
		},
	}), "get#path/123")
}

func TestStorage_GetWhenThenKeyFromRequest(t *testing.T) {
	s := InMemoryStore{}
	assert.Equal(t, s.getWhenThenKeyFromRequest(types.StoreRequest{
		Url:    "/path/123",
		Method: http.MethodGet,
	}), "get#path/123")
}

func TestStorage_Store_and_Get_CleanableMismatches_Match(t *testing.T) {
	t.SkipNow()
	s := InMemoryStore{}
	test := types.WhenThen{
		When: types.When{
			URL:    "abc/def",
			Method: "get",
			Headers: types.CleanHeaders(map[string][]string{
				"accept": {"a", "b"},
			}),
			Body: "abc",
		},
		Then: types.Then{
			Status: 100,
		},
	}

	key, err := s.Store(test)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	actual, err := s.FindByRequest(types.NewStoreRequest("/abc/def", "get", types.Header{
		"Accept": {"a", "b"},
	}, strings.NewReader("A b\nc")))

	assert.NoError(t, err)

	assert.Equal(t, test.Then, *actual)
}

func TestStorage_Store_and_Get_Header_KeyMismatch_NoMatch(t *testing.T) {
	s := InMemoryStore{}
	test := types.WhenThen{
		When: types.When{
			URL:    "abc/def",
			Method: "get",
		},
		Then: types.Then{
			Status: 100,
		},
	}

	key, err := s.Store(test)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	_, err = s.FindByRequest(types.NewStoreRequest("/abc/xyz", "get", nil, nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, NOT_FOUND))
}

func TestStorage_Store_and_Get_Header_ContentMismatch_NoMatch(t *testing.T) {
	s := InMemoryStore{}
	test := types.WhenThen{
		When: types.When{
			URL:     "abc/def",
			Method:  "get",
			Headers: map[string][]string{"something": {"this"}},
		},
		Then: types.Then{
			Status: 100,
			Delay:  1,
			Headers: map[string][]string{
				"accept": {"app/json", "app/xml"},
			},
			Body: "abc",
		},
	}

	key, err := s.Store(test)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	_, err = s.FindByRequest(types.NewStoreRequest("/abc/def", "get", types.Header{"something": {"different"}}, nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, NOT_FOUND))
}

func TestStorage_Store_and_Get_Header_When_IsSubsetOf_HeaderRequest_Match(t *testing.T) {
	s := InMemoryStore{}
	test := types.WhenThen{
		When: types.When{
			URL:     "abc/def",
			Method:  "get",
			Headers: map[string][]string{"something": {"this"}},
		},
		Then: types.Then{
			Status: 100,
			Delay:  1,
			Headers: map[string][]string{
				"accept": {"app/json", "app/xml"},
			},
			Body: "abc",
		},
	}

	key, err := s.Store(test)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	v, err := s.FindByRequest(types.NewStoreRequest("/abc/def", "get", types.Header{"something": {"this", "different"}}, nil))
	assert.NoError(t, err)
	assert.NotNil(t, v)

}

func TestStorage_Store_and_Get_Header_Request_IsSubsetOf_HeaderWhen_NoMatch(t *testing.T) {
	s := InMemoryStore{}
	test := types.WhenThen{
		When: types.When{
			URL:     "abc/def",
			Method:  "get",
			Headers: map[string][]string{"something": {"this", "both"}},
		},
		Then: types.Then{
			Status: 100,
			Delay:  1,
			Headers: map[string][]string{
				"accept": {"app/json", "app/xml"},
			},
			Body: "abc",
		},
	}

	key, err := s.Store(test)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	_, err = s.FindByRequest(types.NewStoreRequest("/abc/def", "get", types.Header{"something": {"this"}}, nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, NOT_FOUND))
}

func TestStorage_Store_and_Get_Header_No_Intersection_Match(t *testing.T) {
	s := InMemoryStore{}
	test := types.WhenThen{
		When: types.When{
			URL:     "abc/def",
			Method:  "get",
			Headers: map[string][]string{"something": {"this", "both"}},
		},
		Then: types.Then{
			Status: 100,
			Delay:  1,
			Headers: map[string][]string{
				"accept": {"app/json", "app/xml"},
			},
			Body: "abc",
		},
	}

	key, err := s.Store(test)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	v, err := s.FindByRequest(types.NewStoreRequest("/abc/def", "get", types.Header{"different": {"content"}}, nil))
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestStorage_Store_and_Get_Body_WhenBody_RequestNoBody_NoMatch(t *testing.T) {
	s := InMemoryStore{}
	test := types.WhenThen{
		When:types.When{
			URL:     "abc/def",
			Method:  "get",
			Headers: map[string][]string{"something": {"this", "both"}},
			Body:    "abc",
		},
		Then: types.Then{
			Status: 100,
		},
	}

	key, err := s.Store(test)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	_, err = s.FindByRequest(types.NewStoreRequest("/abc/def", "get", nil, nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, NOT_FOUND))
}

func TestStorage_Store_and_Get_Body_WhenNoBody_RequestBody_NoMatch(t *testing.T) {
	s := InMemoryStore{}
	test := types.WhenThen{
		When: types.When{
			URL:     "abc/def",
			Method:  "get",
			Headers: map[string][]string{"something": {"this", "both"}},
		},
		Then: types.Then{
			Status: 100,
		},
	}

	key, err := s.Store(test)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	_, err = s.FindByRequest(types.NewStoreRequest("/abc/def", "get", nil, strings.NewReader("anc")))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, NOT_FOUND))
}

func TestStorage_Store_and_Get_Body_Mismatch_NoMatch(t *testing.T) {
	s := InMemoryStore{}
	test := types.WhenThen{
		When: types.When{
			URL:     "abc/def",
			Method:  "get",
			Headers: map[string][]string{"something": {"this", "both"}},
			Body:    "abc",
		},
		Then: types.Then{
			Status: 100,
		},
	}

	key, err := s.Store(test)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	_, err = s.FindByRequest(types.NewStoreRequest("/abc/def", "get", nil, strings.NewReader("anc")))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, NOT_FOUND))
}