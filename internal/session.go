/*
 *     verteilzentrum
 *     Copyright (C) 2020  bn4t
 *
 *     This program is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *
 *     This program is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *
 *     You should have received a copy of the GNU General Public License
 *     along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package internal

import (
	"github.com/emersion/go-smtp"
	"io"
	"io/ioutil"
	"log"
	"net/mail"
	"strings"
)

// A Session is returned after successful login.
type Session struct {
	From   string
	To     string
	List   string
	Prefix string
}

func (s *Session) Mail(from string, _ smtp.MailOptions) error {
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

			// check if the sender is blacklisted
			if StringInSlice(s.From, list.Blacklist) || StringInSlice("*", list.Blacklist) {
				return &smtp.SMTPError{
					Code:         550,
					EnhancedCode: smtp.EnhancedCode{550},
					Message:      "You are blacklisted on this list",
				}
			}

			// check if a whitelist exists and if yes if the sender is whitelisted
			if len(list.Whitelist) > 0 && !StringInSlice(s.From, list.Whitelist) && !StringInSlice("*", list.Whitelist) {
				return &smtp.SMTPError{
					Code:         550,
					EnhancedCode: smtp.EnhancedCode{550},
					Message:      "You are not whitelisted on this list",
				}
			}

			// add a user as a subscriber
			if s.Prefix == "subscribe" {
				err := Subscribe(s.From, s.List)
				if err != nil {
					log.Print("Error while trying to subscribe user " + s.From + " to list " + s.List + ": ")
					log.Print(err)
					return &smtp.SMTPError{
						Code:         451,
						EnhancedCode: smtp.EnhancedCode{451},
						Message:      "Internal server error",
					}
				}

				err = SendSubscribeNotif(s.From, s.List)
				if err != nil {
					log.Print("Error while sending subscribe confirmation to user " + s.From + " for list " + s.List + ": ")
					log.Print(err)
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
				err := Unsubscribe(s.From, s.List)
				if err != nil {
					log.Print("Error while trying to unsubscribe user " + s.From + " from list " + s.List + ": ")
					log.Print(err)
					return &smtp.SMTPError{
						Code:         451,
						EnhancedCode: smtp.EnhancedCode{451},
						Message:      "Internal server error",
					}
				}

				err = SendUnsubscribeNotif(s.From, s.List)
				if err != nil {
					log.Print("Error while sending unsubscribe confirmation to user " + s.From + " for list " + s.List + ": ")
					log.Print(err)
					return &smtp.SMTPError{
						Code:         451,
						EnhancedCode: smtp.EnhancedCode{451},
						Message:      "Internal server error",
					}
				}
				return nil
			}
		}
	}

	return nil
}

func (s *Session) Data(r io.Reader) error {
	// discard data if it is a subscribe or unsubscribe request or if the mail is bounced
	if s.Prefix == "subscribe" || s.Prefix == "unsubscribe" || s.Prefix == "bounce" {
		return nil
	}

	for _, list := range Config.Lists {
		if list.Name == s.List {
			// check if the sender is allowed to publish on this list
			if !StringInSlice(s.From, list.CanPublish) && !StringInSlice("*", list.CanPublish) {
				return &smtp.SMTPError{
					Code:         550,
					EnhancedCode: smtp.EnhancedCode{550},
					Message:      "You are not allowed to publish on this list",
				}
			}
		}
	}

	subs, err := GetSubscribers(s.List)
	if err != nil {
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{451},
			Message:      "Internal server error",
		}
	}

	m, err := mail.ReadMessage(r)
	if err != nil {
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{451},
			Message:      "Internal server error",
		}
	}

	// set mailing list headers
	m.Header["List-Unsubscribe"] = []string{"<mailto:unsubscribe+" + s.List + ">"}
	m.Header["List-Post"] = []string{"<mailto:" + s.List + ">"}
	m.Header["List-Subscribe"] = []string{"<mailto:subscribe+" + s.List + ">"}
	m.Header["Reply-To"] = []string{s.List}
	m.Header["Sender"] = []string{"\"" + strings.Split(s.List, "@")[0] + "\"" + " <" + s.List + ">"}

	// concat all the mail data
	var strData string
	for k, v := range m.Header {
		strData += k + ": " + strings.Join(v, ",") + "\r\n"
	}
	d, err := ioutil.ReadAll(m.Body)
	if err != nil {
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{451},
			Message:      "Internal server error",
		}
	}
	strData += "\r\n" + string(d)

	for _, val := range subs {
		if s.From == val {
			continue
		}

		// try to send the mail to the subscriber. If this fails queue the message for resending.
		if err := SendMail([]byte(strData), "bounce+"+s.List, val); err != nil {
			if err := AddToMsgQueue(val, "bounce+"+s.List, strData); err != nil {
				log.Print("Error adding message to message queue:")
				log.Print(err)
			}
		}
	}

	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}