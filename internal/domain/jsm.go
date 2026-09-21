package domain

import "strings"

const (
	JSMStatusOpen   = "open"
	JSMStatusClosed = "closed"
	JSMStatusAll    = "all"
)

const (
	JSMCategoryOpen   = "OPEN"
	JSMCategoryClosed = "CLOSED"
)

// ServiceDesk is one JSM customer portal desk.
type ServiceDesk struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// RequestType is a customer request type on one desk.
type RequestType struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	DeskID string `json:"desk_id"`
}

// Comment is a JSM customer comment. Customers may only post public: true.
type Comment struct {
	Body   string `json:"body"`
	Public bool   `json:"public"`
}

// CustomerRequest is one customer-portal request.
// PortalURL is https://<hostname>/servicedesk/customer/portal/{desk}/{KEY}.
type CustomerRequest struct {
	Key            string    `json:"key"`
	Site           string    `json:"site"`
	DeskID         string    `json:"desk_id"`
	TypeID         string    `json:"type_id,omitempty"`
	Summary        string    `json:"summary"`
	Description    string    `json:"description,omitempty"`
	Status         string    `json:"status"`
	StatusCategory string    `json:"status_category,omitempty"`
	PortalURL      string    `json:"portal_url"`
	Comments       []Comment `json:"comments,omitempty"`
}

// DeskList is GET /servicedesk for one customer site.
type DeskList struct {
	Site  string        `json:"site"`
	Count int           `json:"count"`
	Items []ServiceDesk `json:"items"`
}

// RequestTypeList is GET /servicedesk/{id}/requesttype.
type RequestTypeList struct {
	DeskID string        `json:"desk_id"`
	Count  int           `json:"count"`
	Items  []RequestType `json:"items"`
}

// RequestList is GET /request filtered by open|closed|all.
type RequestList struct {
	Site   string            `json:"site"`
	Status string            `json:"status"`
	Count  int               `json:"count"`
	Items  []CustomerRequest `json:"items"`
}

// CreateRequest is POST /request. Customers cannot set raiseOnBehalfOf.
type CreateRequest struct {
	DeskID      string
	TypeID      string
	Summary     string
	Description string
}

// PortalURL is https://<hostname>/servicedesk/customer/portal/{desk}/{KEY}.
func PortalURL(hostname, deskID, key string) string {
	return "https://" + hostname + "/servicedesk/customer/portal/" + deskID + "/" + key
}

// NormalizeJSMStatus maps CLI --status open|closed|all. Empty defaults to all.
func NormalizeJSMStatus(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		s = JSMStatusAll
	}
	switch s {
	case JSMStatusOpen, JSMStatusClosed, JSMStatusAll:
		return s, nil
	default:
		return "", Usage("status must be open, closed, or all").WithHint("atlas jsm list --status open")
	}
}

// JSMRequestStatusQuery is the customer REST requestStatus value.
func JSMRequestStatusQuery(s string) string {
	switch s {
	case JSMStatusOpen:
		return "OPEN_REQUESTS"
	case JSMStatusClosed:
		return "CLOSED_REQUESTS"
	default:
		return "ALL_REQUESTS"
	}
}
