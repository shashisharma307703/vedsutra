package handler

import (
	"github.com/shashisharma307703/vedantam/internal/service"
)

type OrgHandler struct {
	svc *service.OrgService
}

func NewOrgHandler(svc *service.OrgService) *OrgHandler {
	return &OrgHandler{svc: svc}
}

// func (h *OrgHandler) Create(w http.ResponseWriter, r *http.Request) {
// 	var params dbgen.UpsertOrganizationParams
// 	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	res, err := h.svc.Create(r.Context(), params)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(res)
// }

// func (h *OrgHandler) Get(w http.ResponseWriter, r *http.Request) {
// 	id, err := uuid.Parse(chi.URLParam(r, "orgId"))
// 	if err != nil {
// 		http.Error(w, "invalid uuid", http.StatusBadRequest)
// 		return
// 	}
// 	res, err := h.svc.Get(r.Context(), id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(res)
// }

// func (h *OrgHandler) Update(w http.ResponseWriter, r *http.Request) {
// 	id, err := uuid.Parse(chi.URLParam(r, "orgId"))
// 	if err != nil {
// 		http.Error(w, "invalid uuid", http.StatusBadRequest)
// 		return
// 	}
// 	var params dbgen.ReplaceOrganizationParams
// 	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	params.OrgID = id
// 	res, err := h.svc.Update(r.Context(), params)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(res)
// }

// func (h *OrgHandler) Patch(w http.ResponseWriter, r *http.Request) {
// 	id, err := uuid.Parse(chi.URLParam(r, "orgId"))
// 	if err != nil {
// 		http.Error(w, "invalid uuid", http.StatusBadRequest)
// 		return
// 	}
// 	var updates map[string]interface{}
// 	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	res, err := h.svc.Patch(r.Context(), id, updates)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(res)
// }

// func (h *OrgHandler) Delete(w http.ResponseWriter, r *http.Request) {
// 	id, err := uuid.Parse(chi.URLParam(r, "orgId"))
// 	if err != nil {
// 		http.Error(w, "invalid uuid", http.StatusBadRequest)
// 		return
// 	}
// 	if err := h.svc.Delete(r.Context(), id); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteHeader(http.StatusNoContent)
// }

// func (h *OrgHandler) List(w http.ResponseWriter, r *http.Request) {
// 	q := r.URL.Query()
// 	limit, _ := strconv.ParseUint(q.Get("limit"), 10, 64)
// 	offset, _ := strconv.ParseUint(q.Get("offset"), 10, 64)

// 	filter := repository.OrgListAndSearchFilter{
// 		City:       q.Get("city"),
// 		SearchTerm: q.Get("search"),
// 		Limit:      limit,
// 		Offset:     offset,
// 	}
// 	res, err := h.svc.List(r.Context(), filter)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(res)
// }
