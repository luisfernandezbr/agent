package server

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/log"
)

func (s *Server) sendSlackMessage(logger sdk.Logger, action string, customerID string, instanceID string, refType string, rerr error, args ...interface{}) {

	errFunc := func(err error) {
		log.Error(logger, "error sending message to slack", "err", err)
	}

	args = append(args, "customer_id", customerID)
	args = append(args, "instance_id", instanceID)
	args = append(args, "ref_type", refType)

	if rerr != nil {
		args = append(args, "error", rerr.Error())
	}

	state, err := s.newState(customerID, instanceID)
	if err != nil {
		errFunc(err)
		return
	}
	var errorMessage string
	exists, err := state.Get("error-message-"+action, &errorMessage)
	if err != nil {
		errFunc(err)
		return
	}

	if exists {
		// there was an error before, but there isn't one now, let's celebrate it's fixed!
		if rerr == nil {
			if err := state.Delete("error-message-" + action); err != nil {
				errFunc(err)
				return
			}
			msg := formatSlackMessage("🎉  *"+refType+"* the previous "+action+" error has been fixed!", args...)
			if err := s.slack.SendMessage(msg); err != nil {
				errFunc(err)
				return
			}
			return
		}
		if rerr.Error() == errorMessage {
			// don't send it again
			return
		}
	}
	if rerr == nil {
		return
	}
	if err := state.Set("error-message-"+action, rerr.Error()); err != nil {
		errFunc(err)
		return
	}
	msg := formatSlackMessage("⚠️ *"+refType+"* error running "+action, args...)
	if err := s.slack.SendMessage(msg); err != nil {
		errFunc(err)
	}
}

func formatSlackMessage(msg string, args ...interface{}) string {

	parts := []string{}
	var val string
	for i, m := range args {
		if i%2 != 0 {
			b, _ := json.Marshal(m)
			val += "=" + string(b)
			parts = append(parts, val)
		} else {
			val = fmt.Sprint(m)
		}
	}
	if len(args)%2 != 0 {
		val += "=(MISSING)"
		parts = append(parts, val)
	}
	sort.Strings(parts)
	cnt := strings.Join(parts, "\n")

	return msg + " ```" + cnt + "```"
}
