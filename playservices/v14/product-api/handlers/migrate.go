package handlers

import (
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v14/product-api/data"
	"go.uber.org/zap"
)

func (p *ProductsHandler) MigrateDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	p.l.Info("Handle GET Products")
	res, err := p.db.MigrateDocs(r.Context())
	if err != nil {
		p.l.Error("error migrating docs", zap.Error(err))
		http.Error(w, "error migrating docs", http.StatusInternalServerError)
	}
	//w.WriteHeader(http.StatusOK)
	err = data.ToJSON(res.InsertedIDs, w)
	if err != nil {
		p.l.Error("error writing", zap.Error(err))
	}
}
