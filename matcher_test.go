package genlib

import (
	"testing"
)

func TestMatcher(t *testing.T) {
	cases := []struct {
		name       string
		matcher    Matcher
		matched    []string
		notMatched []string
	}{
		{
			name:       "name ignored case false",
			matcher:    &stringMatcher{string: "MyName", ignoreCase: false},
			matched:    []string{"MyName"},
			notMatched: []string{"", "myName", "Myname", "myname", "diff"},
		},

		{
			name:       "name ignored case false",
			matcher:    &stringMatcher{string: "MyName", ignoreCase: false},
			matched:    []string{"MyName"},
			notMatched: []string{"", "myName", "Myname", "myname", "diff"},
		},

		{
			name:       "name ignored case true",
			matcher:    &stringMatcher{string: "MyName", ignoreCase: true},
			matched:    []string{"MyName", "myName", "Myname", "myname"},
			notMatched: []string{"myname1", "myname2", "whatever"},
		},

		{
			name:       "regex - suffix",
			matcher:    &regexMatcher{regex: "^.+UseCase$"},
			matched:    []string{"myUseCase", "yourUseCase", "WhateverUseCase"},
			notMatched: []string{"UseCase", "myuseCase", "yourUsecase", "whateverUseCaseStruct"},
		},

		{
			name:       "regex - suffix ignore case",
			matcher:    &regexMatcher{regex: "(?i)^.+UseCase$"},
			matched:    []string{"myUseCase", "myuseCase", "yourUsecase", "yourUseCase", "WhateverUseCase"},
			notMatched: []string{"UseCase", "whateverUseCaseStruct"},
		},

		{
			name: "list - name ignored case false + regex suffix",
			matcher: &listMatcher{matchers: []Matcher{
				&stringMatcher{string: "MyName", ignoreCase: false},
				&regexMatcher{regex: "^.+UseCase$"},
			}},
			matched:    []string{"MyName", "myUseCase", "yourUseCase", "WhateverUseCase"},
			notMatched: []string{"", "myName", "Myname", "myname", "diff", "UseCase", "myuseCase", "yourUsecase", "whateverUseCaseStruct"},
		},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, m := range tc.matched {
				result := tc.matcher.Match(m)
				if !result {
					t.Errorf("expect %s matched but it did not match", m)
				}
			}

			for _, m := range tc.notMatched {
				result := tc.matcher.Match(m)
				if result {
					t.Errorf("expect %s is not matched but it did match", m)
				}
			}
		})
	}
}
