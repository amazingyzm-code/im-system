package api

import (
	"encoding/json"
	"im-system/group"
	"im-system/pkg/db"
	"net/http"
	"strconv"
)

type Server struct {
	groupSvc *group.Service
}

func NewServer(groupSvc *group.Service) *Server {
	return &Server{groupSvc: groupSvc}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/history/single", s.singleHistory)
	mux.HandleFunc("/api/history/group", s.groupHistory)
	mux.HandleFunc("/api/group/join", s.joinGroup)
	mux.HandleFunc("/api/group/leave", s.leaveGroup)
}

// GET /api/history/single?uid1=1&uid2=2&last_id=0&limit=20
func (s *Server) singleHistory(w http.ResponseWriter, r *http.Request) {
	uid1, _ := strconv.ParseInt(r.URL.Query().Get("uid1"), 10, 64)
	uid2, _ := strconv.ParseInt(r.URL.Query().Get("uid2"), 10, 64)
	lastID, _ := strconv.ParseInt(r.URL.Query().Get("last_id"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	msgs, err := db.GetHistory(uid1, uid2, lastID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, msgs)
}

// GET /api/history/group?group_id=1&last_id=0&limit=20
func (s *Server) groupHistory(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.ParseInt(r.URL.Query().Get("group_id"), 10, 64)
	lastID, _ := strconv.ParseInt(r.URL.Query().Get("last_id"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	msgs, err := db.GetGroupHistory(groupID, lastID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, msgs)
}

// POST /api/group/join?group_id=1&uid=2
func (s *Server) joinGroup(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.ParseInt(r.URL.Query().Get("group_id"), 10, 64)
	uid, _ := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 64)
	if err := s.groupSvc.AddMember(groupID, uid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// POST /api/group/leave?group_id=1&uid=2
func (s *Server) leaveGroup(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.ParseInt(r.URL.Query().Get("group_id"), 10, 64)
	uid, _ := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 64)
	if err := s.groupSvc.RemoveMember(groupID, uid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
