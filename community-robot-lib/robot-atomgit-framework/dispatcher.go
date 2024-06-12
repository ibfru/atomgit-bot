package framework

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/opensourceways/community-robot-lib/config"
	sdk "github.com/opensourceways/go-atomgit/atomgit"
	"github.com/sirupsen/logrus"
)

const (
	LogFieldOrg       = "org"
	LogFieldRepo      = "repo"
	LogFieldEventId   = "event_id"
	LogFieldEventType = "event-type"
	LogFieldPayload   = "payload"
	logFieldURL       = "url"
	logFieldAction    = "action"

	UserAgentHeader = "Robot-AtomGit-Access"
)

type dispatcher struct {
	agent *config.ConfigAgent

	h handlers

	// Tracks running handlers for graceful shutdown
	wg sync.WaitGroup

	// secret usage
	hmac func() []byte
}

func (d *dispatcher) Wait() {
	d.wg.Wait() // Handle remaining requests
}

func (d *dispatcher) Dispatch(eventType string, payload []byte, l *logrus.Entry) error {
	hook, err := sdk.ParseWebHook(eventType, payload)
	if err != nil {
		return err
	}

	switch hookType := hook.(type) {
	case *sdk.AccessEvent:
		d.wg.Add(1)
		go d.handleAccessEvent(hookType, l, payload)
	case *sdk.IssuesEvent:
		d.wg.Add(1)
		go d.handleIssueEvent(hookType, l)
	case *sdk.PullRequestEvent:
		d.wg.Add(1)
		go d.handlePullRequestEvent(hookType, l)
	case *sdk.PushEvent:
		d.wg.Add(1)
		go d.handlePushEvent(hookType, l)
	case *sdk.IssueCommentEvent:
		d.wg.Add(1)
		go d.handleIssueCommentEvent(hookType, l)
	case *sdk.PullRequestReviewEvent:
		d.wg.Add(1)
		go d.handleReviewEvent(hookType, l)
	case *sdk.PullRequestReviewCommentEvent:
		d.wg.Add(1)
		go d.handleReviewCommentEvent(hookType, l)
	default:
		l.Debug("Ignoring unknown event type")
	}

	return nil
}

func (d *dispatcher) getConfig() config.Config {
	_, c := d.agent.GetConfig()

	return c
}

// handleAccessEvent access robot handle request that come form webhook
func (d *dispatcher) handleAccessEvent(e *sdk.AccessEvent, l *logrus.Entry, payload []byte) {
	defer d.wg.Done()

	org, repo := e.GetRepo().GetOrgAndRepo()

	l = l.WithFields(logrus.Fields{
		LogFieldOrg:  org,
		LogFieldRepo: repo,
	})

	if err := d.h.accessHandlers(e, d.getConfig(), l, payload); err != nil {
		l.WithError(err).Error()
	} else {
		l.Info()
	}
}

func (d *dispatcher) handleIssueEvent(e *sdk.IssuesEvent, l *logrus.Entry) {
	defer d.wg.Done()

	l = l.WithFields(logrus.Fields{
		logFieldURL:    e.GetIssue().GetHTMLURL(),
		logFieldAction: e.GetAction(),
	})

	if err := d.h.issueHandlers(e, d.getConfig(), l); err != nil {
		l.WithError(err).Error()
	} else {
		l.Info()
	}
}

func (d *dispatcher) handlePullRequestEvent(e *sdk.PullRequestEvent, l *logrus.Entry) {
	defer d.wg.Done()

	l = l.WithFields(logrus.Fields{
		logFieldURL:    e.GetPullRequest().GetHTMLURL(),
		logFieldAction: e.GetAction(),
	})

	if err := d.h.pullRequestHandler(e, d.getConfig(), l); err != nil {
		l.WithError(err).Error()
	} else {
		l.Info()
	}
}

func (d *dispatcher) handlePushEvent(e *sdk.PushEvent, l *logrus.Entry) {
	defer d.wg.Done()
	l = l.WithFields(logrus.Fields{
		LogFieldOrg:  e.GetRepo().GetOwner().GetLogin(),
		LogFieldRepo: e.GetRepo().GetName(),
		"ref":        e.GetRef(),
		"head":       e.GetAfter(),
	})

	if err := d.h.pushEventHandler(e, d.getConfig(), l); err != nil {
		l.WithError(err).Error()
	} else {
		l.Info()
	}
}

func (d *dispatcher) handleIssueCommentEvent(e *sdk.IssueCommentEvent, l *logrus.Entry) {
	defer d.wg.Done()

	l = l.WithFields(logrus.Fields{
		logFieldURL:    e.GetIssue().GetHTMLURL(),
		logFieldAction: e.GetAction(),
	})

	if err := d.h.issueCommentHandler(e, d.getConfig(), l); err != nil {
		l.WithError(err).Error()
	} else {
		l.Info()
	}
}

func (d *dispatcher) handleReviewEvent(e *sdk.PullRequestReviewEvent, l *logrus.Entry) {
	defer d.wg.Done()

	org, repo := e.GetRepo().GetOrgAndRepo()
	l = l.WithFields(logrus.Fields{
		LogFieldOrg:  org,
		LogFieldRepo: repo,
		"review":     e.GetReview().GetID(),
		"reviewer":   e.GetReview().GetUser().GetLogin(),
		"url":        e.GetReview().GetHTMLURL(),
	})

	if err := d.h.reviewEventHandler(e, d.getConfig(), l); err != nil {
		l.WithError(err).Error()
	} else {
		l.Info()
	}
}

func (d *dispatcher) handleReviewCommentEvent(e *sdk.PullRequestReviewCommentEvent, l *logrus.Entry) {
	defer d.wg.Done()

	org, repo := e.GetRepo().GetOrgAndRepo()
	l = l.WithFields(logrus.Fields{
		LogFieldOrg:  org,
		LogFieldRepo: repo,
		"review":     e.GetComment().GetPullRequestReviewID(),
		"reviewer":   e.GetComment().GetUser().GetLogin(),
		"url":        e.GetComment().GetHTMLURL(),
	})

	if err := d.h.reviewCommentEventHandler(e, d.getConfig(), l); err != nil {
		l.WithError(err).Error()
	} else {
		l.Info()
	}
}

func (d *dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	eventType, eventGUID, payload, ok := parseRequest(w, r, d.hmac)
	if !ok {
		return
	}

	evt := eventType
	if strings.HasPrefix(eventType, sdk.EventCustomToAccess) {
		evt = sdk.EventCustomToAccess
		eventType = eventType[10:]
	}

	l := logrus.WithFields(
		logrus.Fields{
			LogFieldEventType: eventType,
			LogFieldEventId:   eventGUID,
		},
	)

	if err := d.Dispatch(evt, payload, l); err != nil {
		l.WithError(err).Error()
	}
}

func parseRequest(w http.ResponseWriter, r *http.Request, getHmac func() []byte) (eventType string, uuid string, payload []byte, ok bool) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logrus.Warn("when request body close, error occurred:", err)
		}
	}(r.Body)

	resp := func(code int, msg string) {
		http.Error(w, msg, code)
	}

	if eventType = r.Header.Get("X-AtomGit-Event"); eventType == "" {
		resp(http.StatusBadRequest, "400 Bad Request: Missing X-AtomGit-Event Header")
		return
	}

	v, err := io.ReadAll(r.Body)
	if err != nil {
		resp(http.StatusInternalServerError, "500 Internal Server Error: Failed to read request body")
		return
	}

	ua := r.Header.Get("User-Agent")
	if ua == "AtomGit-Hookshot" {
		// add header for Access Bot
		eventType = sdk.EventCustomToAccess + eventType

		if uuid = r.Header.Get("X-AtomGit-Delivery"); uuid == "" {
			resp(http.StatusBadRequest, "400 Bad Request: Missing X-AtomGit-Delivery Header")
			return
		}

		sign := r.Header.Get("X-Hub-Signature-256")
		if sign == "" || !strings.HasPrefix(sign, "sha256=") {
			resp(http.StatusForbidden, "403 Forbidden: Missing X-Hub-Signature-256 Header")
			return
		}
		// Validate the payload with our HMAC secret.
		if !hmac.Equal([]byte(sign[7:]), []byte(payloadSignature(v, getHmac()))) {
			resp(http.StatusForbidden, "403 Forbidden: Invalid X-Hub-Signature-256")
			return
		}

		resp(http.StatusOK, "The request was accepted by access's robot, inform to webhook.")

	} else {
		if ua != UserAgentHeader {
			resp(http.StatusBadRequest, "400 Bad Request: unknown User-Agent Header")
			return
		}
	}

	payload = v
	ok = true

	return
}

func payloadSignature(payload []byte, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
