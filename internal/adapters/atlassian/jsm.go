package atlassian

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// JSMAPI is the in-process Garda customer adapter used by ATLAS_FAKE and tests.
type JSMAPI struct {
	Memory *Memory
}

func (j JSMAPI) Desks(ctx context.Context, hostname string) (domain.DeskList, error) {
	if j.Memory == nil {
		return domain.DeskList{}, domain.Service("jsm memory not configured")
	}
	return j.Memory.ListDesks(ctx, hostname)
}

func (j JSMAPI) Types(ctx context.Context, hostname, deskID string) (domain.RequestTypeList, error) {
	if j.Memory == nil {
		return domain.RequestTypeList{}, domain.Service("jsm memory not configured")
	}
	return j.Memory.ListTypes(ctx, hostname, deskID)
}

func (j JSMAPI) List(ctx context.Context, hostname, status string) (domain.RequestList, error) {
	if j.Memory == nil {
		return domain.RequestList{}, domain.Service("jsm memory not configured")
	}
	return j.Memory.ListRequests(ctx, hostname, status)
}

func (j JSMAPI) Get(ctx context.Context, hostname, key string) (domain.CustomerRequest, error) {
	if j.Memory == nil {
		return domain.CustomerRequest{}, domain.Service("jsm memory not configured")
	}
	return j.Memory.GetRequest(ctx, hostname, key)
}

func (j JSMAPI) Create(ctx context.Context, hostname string, in domain.CreateRequest, dryRun bool) (domain.CustomerRequest, error) {
	if j.Memory == nil {
		return domain.CustomerRequest{}, domain.Service("jsm memory not configured")
	}
	return j.Memory.CreateRequest(ctx, hostname, in, dryRun)
}

func (j JSMAPI) Comment(ctx context.Context, hostname, key, body string, dryRun bool) error {
	if j.Memory == nil {
		return domain.Service("jsm memory not configured")
	}
	return j.Memory.CommentRequest(ctx, hostname, key, body, dryRun)
}

func (j JSMAPI) Transition(ctx context.Context, hostname, key, id string, dryRun bool) (domain.CustomerRequest, error) {
	if j.Memory == nil {
		return domain.CustomerRequest{}, domain.Service("jsm memory not configured")
	}
	return j.Memory.TransitionRequest(ctx, hostname, key, id, dryRun)
}

func (m *Memory) ListDesks(_ context.Context, hostname string) (domain.DeskList, error) {
	if m == nil {
		return domain.DeskList{}, domain.Service("jsm memory not configured")
	}
	hostname, err := requireGardaHost(hostname)
	if err != nil {
		return domain.DeskList{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	items := append([]domain.ServiceDesk(nil), m.desks...)
	sort.Slice(items, func(i, j int) bool { return deskOrder(items[i].ID) < deskOrder(items[j].ID) })
	return domain.DeskList{Site: hostname, Count: len(items), Items: items}, nil
}

func (m *Memory) ListTypes(_ context.Context, hostname, deskID string) (domain.RequestTypeList, error) {
	if m == nil {
		return domain.RequestTypeList{}, domain.Service("jsm memory not configured")
	}
	if _, err := requireGardaHost(hostname); err != nil {
		return domain.RequestTypeList{}, err
	}
	deskID = strings.TrimSpace(deskID)
	if deskID == "" {
		return domain.RequestTypeList{}, domain.Usage("desk is required").WithHint("atlas jsm types --desk 3")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.hasDeskLocked(deskID) {
		return domain.RequestTypeList{}, domain.NotFound("service desk not found").WithHint("atlas jsm desks")
	}
	items := append([]domain.RequestType(nil), m.types[deskID]...)
	return domain.RequestTypeList{DeskID: deskID, Count: len(items), Items: items}, nil
}

func (m *Memory) ListRequests(_ context.Context, hostname, status string) (domain.RequestList, error) {
	if m == nil {
		return domain.RequestList{}, domain.Service("jsm memory not configured")
	}
	hostname, err := requireGardaHost(hostname)
	if err != nil {
		return domain.RequestList{}, err
	}
	status, err = domain.NormalizeJSMStatus(status)
	if err != nil {
		return domain.RequestList{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]domain.CustomerRequest, 0)
	for _, req := range m.requests {
		if req.Site != hostname {
			continue
		}
		if status == domain.JSMStatusOpen && req.StatusCategory != domain.JSMCategoryOpen {
			continue
		}
		if status == domain.JSMStatusClosed && req.StatusCategory != domain.JSMCategoryClosed {
			continue
		}
		items = append(items, cloneRequest(req))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
	return domain.RequestList{Site: hostname, Status: status, Count: len(items), Items: items}, nil
}

func (m *Memory) GetRequest(_ context.Context, hostname, key string) (domain.CustomerRequest, error) {
	if m == nil {
		return domain.CustomerRequest{}, domain.Service("jsm memory not configured")
	}
	hostname, err := requireGardaHost(hostname)
	if err != nil {
		return domain.CustomerRequest{}, err
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	if key == "" {
		return domain.CustomerRequest{}, domain.Usage("request key is required").WithHint("atlas jsm get EOS-1")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	req, ok := m.requests[memKey(hostname, key)]
	if !ok {
		return domain.CustomerRequest{}, domain.NotFound("request not found").WithHint("check the key; use atlas jsm, not atlas jira get")
	}
	return cloneRequest(req), nil
}

func (m *Memory) CreateRequest(_ context.Context, hostname string, in domain.CreateRequest, dryRun bool) (domain.CustomerRequest, error) {
	if m == nil {
		return domain.CustomerRequest{}, domain.Service("jsm memory not configured")
	}
	hostname, err := requireGardaHost(hostname)
	if err != nil {
		return domain.CustomerRequest{}, err
	}
	deskID := strings.TrimSpace(in.DeskID)
	typeID := strings.TrimSpace(in.TypeID)
	summary := strings.TrimSpace(in.Summary)
	if deskID == "" || typeID == "" || summary == "" {
		return domain.CustomerRequest{}, domain.Usage("create requires --desk, --type, and --summary").WithHint("atlas jsm create --desk 3 --type 40 --summary '…'")
	}
	fields := map[string]any{"summary": summary}
	if strings.TrimSpace(in.Description) != "" {
		fields["description"] = in.Description
	}
	// Customer REST: never raiseOnBehalfOf or requestParticipants.
	payload := map[string]any{
		"serviceDeskId":      deskID,
		"requestTypeId":      typeID,
		"requestFieldValues": fields,
	}
	preview := domain.CustomerRequest{
		Site:           hostname,
		DeskID:         deskID,
		TypeID:         typeID,
		Summary:        summary,
		Description:    in.Description,
		Status:         "Waiting for support",
		StatusCategory: domain.JSMCategoryOpen,
	}
	if dryRun {
		m.mu.Lock()
		m.lastCreate = payload
		m.mu.Unlock()
		return preview, nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	desk, ok := m.deskLocked(deskID)
	if !ok {
		return domain.CustomerRequest{}, domain.NotFound("service desk not found").WithHint("atlas jsm desks")
	}
	if !m.hasTypeLocked(deskID, typeID) {
		return domain.CustomerRequest{}, domain.NotFound("request type not found").WithHint("atlas jsm types --desk " + deskID)
	}
	n := m.nextReq[desk.Key]
	if n == 0 {
		n = 1
	}
	preview.Key = fmt.Sprintf("%s-%d", desk.Key, n)
	preview.PortalURL = domain.PortalURL(hostname, deskID, preview.Key)
	m.nextReq[desk.Key] = n + 1
	m.putRequestLocked(preview)
	m.jsmTransitions[memKey(hostname, preview.Key)] = []string{"21"}
	m.lastCreate = payload
	return cloneRequest(preview), nil
}

func (m *Memory) CommentRequest(_ context.Context, hostname, key, body string, dryRun bool) error {
	if m == nil {
		return domain.Service("jsm memory not configured")
	}
	hostname, err := requireGardaHost(hostname)
	if err != nil {
		return err
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	body = strings.TrimSpace(body)
	if key == "" {
		return domain.Usage("request key is required").WithHint("atlas jsm comment EOS-1 --body '…'")
	}
	if body == "" {
		return domain.Usage("comment requires --body").WithHint("atlas jsm comment EOS-1 --body '…'")
	}
	comment := domain.Comment{Body: body, Public: true}
	m.mu.Lock()
	defer m.mu.Unlock()
	req, ok := m.requests[memKey(hostname, key)]
	if !ok {
		return domain.NotFound("request not found").WithHint("check the key; use atlas jsm, not atlas jira comment")
	}
	m.lastComment = comment
	if dryRun {
		return nil
	}
	req.Comments = append(append([]domain.Comment(nil), req.Comments...), comment)
	m.putRequestLocked(req)
	return nil
}

func (m *Memory) TransitionRequest(_ context.Context, hostname, key, id string, dryRun bool) (domain.CustomerRequest, error) {
	if m == nil {
		return domain.CustomerRequest{}, domain.Service("jsm memory not configured")
	}
	hostname, err := requireGardaHost(hostname)
	if err != nil {
		return domain.CustomerRequest{}, err
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	id = strings.TrimSpace(id)
	if key == "" {
		return domain.CustomerRequest{}, domain.Usage("request key is required").WithHint("atlas jsm transition EOS-1 --id 21")
	}
	if id == "" {
		return domain.CustomerRequest{}, domain.Usage("transition requires --id").WithHint("atlas jsm transition EOS-1 --id 21")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	req, ok := m.requests[memKey(hostname, key)]
	if !ok {
		return domain.CustomerRequest{}, domain.NotFound("request not found").WithHint("check the key")
	}
	ids := m.jsmTransitions[memKey(hostname, key)]
	if !containsFold(ids, id) {
		return domain.CustomerRequest{}, domain.Usagef("unknown transition %q", id).WithHint("available: " + strings.Join(ids, ", "))
	}
	next := cloneRequest(req)
	next.Status = "Closed"
	next.StatusCategory = domain.JSMCategoryClosed
	if dryRun {
		return next, nil
	}
	m.putRequestLocked(next)
	return cloneRequest(next), nil
}

// RequestCount is the seeded plus persisted customer-request count (tests).
func (m *Memory) RequestCount() int {
	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// LastComment is the last comment payload (always public: true).
func (m *Memory) LastComment() domain.Comment {
	if m == nil {
		return domain.Comment{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastComment
}

// LastCreatePayload is the last POST /request body. Must not include raiseOnBehalfOf.
func (m *Memory) LastCreatePayload() map[string]any {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lastCreate == nil {
		return nil
	}
	out := make(map[string]any, len(m.lastCreate))
	for k, v := range m.lastCreate {
		out[k] = v
	}
	return out
}

// SearchCalls is how many times licensed Jira Search ran on this seed (tests).
func (m *Memory) SearchCalls() int {
	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.searchCalls
}

func (m *Memory) putRequest(req domain.CustomerRequest) {
	m.putRequestLocked(req)
}

func (m *Memory) putRequestLocked(req domain.CustomerRequest) {
	if req.PortalURL == "" {
		req.PortalURL = domain.PortalURL(req.Site, req.DeskID, req.Key)
	}
	m.requests[memKey(req.Site, req.Key)] = req
}

func (m *Memory) hasDeskLocked(id string) bool {
	_, ok := m.deskLocked(id)
	return ok
}

func (m *Memory) deskLocked(id string) (domain.ServiceDesk, bool) {
	for _, d := range m.desks {
		if d.ID == id {
			return d, true
		}
	}
	return domain.ServiceDesk{}, false
}

func (m *Memory) hasTypeLocked(deskID, typeID string) bool {
	for _, t := range m.types[deskID] {
		if t.ID == typeID {
			return true
		}
	}
	return false
}

func cloneRequest(req domain.CustomerRequest) domain.CustomerRequest {
	out := req
	if req.Comments != nil {
		out.Comments = append([]domain.Comment(nil), req.Comments...)
	}
	return out
}

func requireGardaHost(hostname string) (string, error) {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		hostname = domain.GardaHostname
	}
	if !strings.EqualFold(hostname, domain.GardaHostname) {
		return "", domain.Usage("jsm is Garda customer REST only").WithHint("sesami-io portal/1 is deferred; SDP stays atlas jira")
	}
	return domain.GardaHostname, nil
}

func deskOrder(id string) int {
	switch id {
	case "3":
		return 0
	case "12":
		return 1
	case "2586":
		return 2
	default:
		return 100
	}
}
