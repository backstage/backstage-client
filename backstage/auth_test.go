package main

import (
	"net/http"
	"os"
	"path"

	ttesting "github.com/tsuru/tsuru/cmd/testing"
	"github.com/tsuru/tsuru/fs/testing"
	. "gopkg.in/check.v1"
)

func (s *S) TestLogin(c *C) {
	rfs := &testing.RecordingFs{FileContent: "current: backstage\noptions:\n  backstage: http://www.example.com"}
	fsystem = rfs
	defer func() {
		fsystem = nil
	}()
	transport := ttesting.Transport{
		Status:  http.StatusOK,
		Message: `{"token_type": "Token", "token": "zyz"}`,
	}
	auth := &Auth{
		client: NewClient(&http.Client{Transport: &transport}),
	}
	r := auth.Login("alice@example.org", "123")
	c.Assert(r, Equals, "Welcome! You've signed in successfully.")
}

func (s *S) TestLoginWithInvalidCredentials(c *C) {
	rfs := &testing.RecordingFs{FileContent: "current: backstage\noptions:\n  backstage: http://www.example.com"}
	fsystem = rfs
	defer func() {
		fsystem = nil
	}()
	transport := ttesting.Transport{
		Status:  http.StatusBadRequest,
		Message: `{"status_code":400,"message":"Invalid Username or Password."}`,
	}
	auth := &Auth{
		client: NewClient(&http.Client{Transport: &transport}),
	}
	r := auth.Login("alice@example.org", "123")
	c.Assert(r, Equals, "Invalid Username or Password.")
}

func (s *S) TestLoginWithInvalidPayload(c *C) {
	rfs := &testing.RecordingFs{FileContent: "current: backstage\noptions:\n  backstage: http://www.example.com"}
	fsystem = rfs
	defer func() {
		fsystem = nil
	}()
	transport := ttesting.Transport{
		Status:  http.StatusBadRequest,
		Message: `{"status_code":400,"message":"The request was bad-formed."}`,
	}
	auth := &Auth{
		client: NewClient(&http.Client{Transport: &transport}),
	}
	r := auth.Login("alice@example.org", "123")
	c.Assert(r, Equals, "The request was bad-formed.")
}

func (s *S) TestLoginWithoutTarget(c *C) {
	rfs := &testing.RecordingFs{}
	fsystem = rfs
	defer func() {
		fsystem = nil
	}()

	auth := &Auth{
		client: NewClient(&http.Client{}),
	}
	r := auth.Login("alice@example.org", "123")
	c.Assert(r, Equals, "You have not selected any target as default. For more details, please run `backstage target-set -h`.")
}

func (s *S) TestLogout(c *C) {
	rfs := &testing.RecordingFs{}
	fsystem = rfs
	defer func() {
		fsystem = nil
	}()
	auth := &Auth{}
	result := auth.Logout()
	c.Assert(result, Equals, "You have signed out successfully.")
	filePath := path.Join(os.ExpandEnv("${HOME}"), ".backstage_token")
	c.Assert(rfs.HasAction("remove "+filePath), Equals, true)
}

func (s *S) TestLogoutWhenNotSignedIn(c *C) {
	auth := &Auth{}
	result := auth.Logout()
	c.Assert(result, Equals, "You are not signed in.")
}

func (s *S) TestReadToken(c *C) {
	rfs := &testing.RecordingFs{FileContent: "Token xyz"}
	fsystem = rfs
	defer func() {
		fsystem = nil
	}()
	token, err := ReadToken()
	c.Assert(err, IsNil)
	c.Assert(token, Equals, "Token xyz")
	filePath := path.Join(os.ExpandEnv("${HOME}"), ".backstage_token")
	c.Assert(rfs.HasAction("openfile "+filePath+" with mode 0600"), Equals, true)
}

func (s *S) TestReadTokenWhenFileNotFound(c *C) {
	_, err := ReadToken()
	c.Assert(err, Not(IsNil))
}

func (s *S) TestDeleteToken(c *C) {
	rfs := &testing.RecordingFs{}
	fsystem = rfs
	defer func() {
		fsystem = nil
	}()
	err := DeleteToken()
	c.Assert(err, IsNil)
	filePath := path.Join(os.ExpandEnv("${HOME}"), ".backstage_token")
	c.Assert(rfs.HasAction("remove "+filePath), Equals, true)
}

func (s *S) TestDeleteTokenWhenFileNotFound(c *C) {
	err := DeleteToken()
	c.Assert(err, Not(IsNil))
}