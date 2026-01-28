package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *MCPServer) handleSubscribe(
	ctx context.Context,
	id any,
	request mcp.SubscribeRequest,
) (*mcp.EmptyResult, *requestError) {
	uri := strings.TrimSpace(request.Params.URI)
	if uri == "" {
		return nil, &requestError{
			id:   id,
			code: mcp.INVALID_PARAMS,
			err:  fmt.Errorf("uri is required"),
		}
	}
	session := ClientSessionFromContext(ctx)
	if session == nil {
		return nil, &requestError{
			id:   id,
			code: mcp.INVALID_REQUEST,
			err:  ErrSessionNotInitialized,
		}
	}

	s.addSubscription(session.SessionID(), uri)
	return &mcp.EmptyResult{}, nil
}

func (s *MCPServer) handleUnsubscribe(
	ctx context.Context,
	id any,
	request mcp.UnsubscribeRequest,
) (*mcp.EmptyResult, *requestError) {
	uri := strings.TrimSpace(request.Params.URI)
	if uri == "" {
		return nil, &requestError{
			id:   id,
			code: mcp.INVALID_PARAMS,
			err:  fmt.Errorf("uri is required"),
		}
	}
	session := ClientSessionFromContext(ctx)
	if session == nil {
		return nil, &requestError{
			id:   id,
			code: mcp.INVALID_REQUEST,
			err:  ErrSessionNotInitialized,
		}
	}

	s.removeSubscription(session.SessionID(), uri)
	return &mcp.EmptyResult{}, nil
}

func (s *MCPServer) addSubscription(sessionID, uri string) {
	if sessionID == "" || uri == "" {
		return
	}
	s.subscriptionsMu.Lock()
	defer s.subscriptionsMu.Unlock()

	set := s.subscriptions[sessionID]
	if set == nil {
		set = make(map[string]struct{})
		s.subscriptions[sessionID] = set
	}
	set[uri] = struct{}{}
}

func (s *MCPServer) removeSubscription(sessionID, uri string) {
	if sessionID == "" || uri == "" {
		return
	}
	s.subscriptionsMu.Lock()
	defer s.subscriptionsMu.Unlock()

	set := s.subscriptions[sessionID]
	if set == nil {
		return
	}
	delete(set, uri)
	if len(set) == 0 {
		delete(s.subscriptions, sessionID)
	}
}

func (s *MCPServer) removeAllSubscriptions(sessionID string) {
	if sessionID == "" {
		return
	}
	s.subscriptionsMu.Lock()
	delete(s.subscriptions, sessionID)
	s.subscriptionsMu.Unlock()
}

func (s *MCPServer) hasSubscriptions(uri string) bool {
	if uri == "" {
		return false
	}
	s.subscriptionsMu.RLock()
	defer s.subscriptionsMu.RUnlock()
	for _, set := range s.subscriptions {
		if subscriptionMatches(set, uri) {
			return true
		}
	}
	return false
}

func (s *MCPServer) SendResourceUpdated(uri string) {
	if uri == "" {
		return
	}
	subscribers := s.subscribedSessions(uri)
	if len(subscribers) == 0 {
		return
	}
	s.sendResourceUpdated(subscribers, uri, nil)
}

// SendResourceUpdatedWithParams sends a resource updated notification with extra fields.
func (s *MCPServer) SendResourceUpdatedWithParams(uri string, params map[string]any) {
	if uri == "" {
		return
	}
	subscribers := s.subscribedSessions(uri)
	if len(subscribers) == 0 {
		return
	}
	s.sendResourceUpdated(subscribers, uri, params)
}

func (s *MCPServer) sendResourceUpdated(subscribers []ClientSession, uri string, params map[string]any) {
	fields := map[string]any{"uri": uri}
	for key, value := range params {
		fields[key] = value
	}
	notification := mcp.JSONRPCNotification{
		JSONRPC: mcp.JSONRPC_VERSION,
		Notification: mcp.Notification{
			Method: mcp.MethodNotificationResourceUpdated,
			Params: mcp.NotificationParams{
				AdditionalFields: fields,
			},
		},
	}
	for _, session := range subscribers {
		_ = s.sendNotificationToSpecificClient(session, notification)
	}
}

func (s *MCPServer) subscribedSessions(uri string) []ClientSession {
	if uri == "" {
		return nil
	}
	s.subscriptionsMu.RLock()
	sessionIDs := make([]string, 0, len(s.subscriptions))
	for sessionID, set := range s.subscriptions {
		if subscriptionMatches(set, uri) {
			sessionIDs = append(sessionIDs, sessionID)
		}
	}
	s.subscriptionsMu.RUnlock()
	if len(sessionIDs) == 0 {
		return nil
	}

	subscribers := make([]ClientSession, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		if sessionValue, ok := s.sessions.Load(sessionID); ok {
			if session, ok := sessionValue.(ClientSession); ok && session.Initialized() {
				subscribers = append(subscribers, session)
			}
		}
	}
	return subscribers
}

// SubscribedURIs returns the unique set of subscribed resource URIs.
func (s *MCPServer) SubscribedURIs() []string {
	s.subscriptionsMu.RLock()
	defer s.subscriptionsMu.RUnlock()
	if len(s.subscriptions) == 0 {
		return nil
	}
	unique := make(map[string]struct{})
	for _, set := range s.subscriptions {
		for uri := range set {
			unique[uri] = struct{}{}
		}
	}
	out := make([]string, 0, len(unique))
	for uri := range unique {
		out = append(out, uri)
	}
	return out
}

func subscriptionMatches(set map[string]struct{}, uri string) bool {
	for sub := range set {
		if uri == sub {
			return true
		}
		if strings.HasPrefix(uri, sub) {
			if len(uri) == len(sub) {
				return true
			}
			if sub != "" && uri[len(sub)] == '/' {
				return true
			}
		}
	}
	return false
}
