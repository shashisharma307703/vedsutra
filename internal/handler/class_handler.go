package handler

import (
	"github.com/shashisharma307703/vedantam/internal/service"
)

type ClassHandler struct {
	svc *service.ClassService
}

func NewClassHandler(svc *service.ClassService) *ClassHandler {
	return &ClassHandler{svc: svc}
}

// func (h *ClassHandler) Create(w http.ResponseWriter, r *http.Request) {
// 	var params dbgen.UpsertClassParams
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

// func (h *ClassHandler) Get(w http.ResponseWriter, r *http.Request) {
// 	orgID, errOrg := uuid.Parse(chi.URLParam(r, "orgId"))
// 	classID, errCls := uuid.Parse(chi.URLParam(r, "classId"))
// 	if errOrg != nil || errCls != nil {
// 		http.Error(w, "invalid configurations", http.StatusBadRequest)
// 		return
// 	}

// 	res, err := h.svc.Get(r.Context(), orgID, classID)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(res)
// }

// func (h *ClassHandler) Update(w http.ResponseWriter, r *http.Request) {
// 	//orgID, errOrg := uuid.Parse(chi.URLParam(r, "orgId"))
// 	classID, errCls := uuid.Parse(chi.URLParam(r, "classId"))
// 	if errCls != nil {
// 		http.Error(w, "invalid request configurations", http.StatusBadRequest)
// 		return
// 	}

// 	var params dbgen.ReplaceClassParams
// 	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	params.ClassLevelID = classID
// 	//params.OrgID = orgID

// 	res, err := h.svc.Update(r.Context(), params)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(res)
// }

// func (h *ClassHandler) Patch(w http.ResponseWriter, r *http.Request) {
// 	orgID, errOrg := uuid.Parse(chi.URLParam(r, "orgId"))
// 	classID, errCls := uuid.Parse(chi.URLParam(r, "classId"))
// 	if errOrg != nil || errCls != nil {
// 		http.Error(w, "bad request parameter identifiers", http.StatusBadRequest)
// 		return
// 	}

// 	var updates map[string]interface{}
// 	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	res, err := h.svc.Patch(r.Context(), orgID, classID, updates)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(res)
// }

// func (h *ClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
// 	orgID, errOrg := uuid.Parse(chi.URLParam(r, "orgId"))
// 	classID, errCls := uuid.Parse(chi.URLParam(r, "classId"))
// 	if errOrg != nil || errCls != nil {
// 		http.Error(w, "bad keys context passed", http.StatusBadRequest)
// 		return
// 	}

// 	if err := h.svc.Delete(r.Context(), orgID, classID); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteHeader(http.StatusNoContent)
// }

// func (h *ClassHandler) List(w http.ResponseWriter, r *http.Request) {
// 	orgID, errOrg := uuid.Parse(chi.URLParam(r, "orgId"))
// 	if errOrg != nil {
// 		http.Error(w, "tenant org missing", http.StatusBadRequest)
// 		return
// 	}

// 	q := r.URL.Query()
// 	limit, _ := strconv.ParseUint(q.Get("limit"), 10, 64)
// 	offset, _ := strconv.ParseUint(q.Get("offset"), 10, 64)

// 	var activePtr *bool
// 	if val := q.Get("is_active"); val != "" {
// 		b, err := strconv.ParseBool(val)
// 		if err == nil {
// 			activePtr = &b
// 		}
// 	}

// 	filter := repository.ClassListAndSearchFilter{
// 		OrgID:      orgID,
// 		IsActive:   activePtr,
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
