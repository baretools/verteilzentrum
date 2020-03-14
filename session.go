package main

import (
	"github.com/emersion/go-smtp"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

// A Session is returned after successful login.
type Session struct {
	From   string
	To     string
	List   string
	Prefix string
}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	if ml := strings.Split(to, "+"); len(ml) > 1 {
		s.Prefix = ml[0]
		s.List = ml[1]
	} else {
		s.List = ml[0]
	}

	// check if the list exists
	if !ListExists(s.List) {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{550},
			Message:      "Specified list is not available",
		}
	}

	for _, list := range Config.Lists {
		if list.Name == s.List {

			// add a user as a subscriber
			if s.Prefix == "subscribe" {
				err := Subscribe(s.From, s.List)
				if err != nil {
					return &smtp.SMTPError{
						Code:         451,
						EnhancedCode: smtp.EnhancedCode{451},
						Message:      "Internal server error",
					}
				}
				return nil
			}

			// unsubscribe a user
			if s.Prefix == "unsubscribe" {
				err := Unubscribe(s.From, s.List)
				if err != nil {
					return &smtp.SMTPError{
						Code:         451,
						EnhancedCode: smtp.EnhancedCode{451},
						Message:      "Internal server error",
					}
				}
				return nil
			}

			// check if the sender is blacklisted
			if StringInSlice(s.From, list.Blacklist) {
				return &smtp.SMTPError{
					Code:         550,
					EnhancedCode: smtp.EnhancedCode{550},
					Message:      "You are blacklisted on this list",
				}
			}

			// check if a whitelist exists and if yes if the sender is whitelisted
			if len(list.Whitelist) > 0 && !StringInSlice(s.From, list.Whitelist) {
				return &smtp.SMTPError{
					Code:         550,
					EnhancedCode: smtp.EnhancedCode{550},
					Message:      "You are not whitelisted on this list",
				}
			}
		}
	}

	return nil
}

func (s *Session) Data(r io.Reader) error {
	if s.Prefix == "subscribe" || s.Prefix == "unsubscribe" {
		return nil
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{451},
			Message:      "Internal server error",
		}
	}

	log.Print(string(data))

	subs, err := GetSubscribers(s.List)
	if err != nil {
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{451},
			Message:      "Internal server error",
		}
	}

	for _, val := range subs {
		// TODO: retry at a later time again if an error occurs
		log.Print(ForwardMail(data, s.From, val))
	}

	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}