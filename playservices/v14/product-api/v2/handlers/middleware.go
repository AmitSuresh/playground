package handlers

import (
	"fmt"
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v14/product-api/v2/data"
	"go.uber.org/zap"
)

// MiddlewareValidateProduct validates the product in the request and calls next if ok
func (p *ProductsHandler) MiddlewareValidateProduct(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			prod := &data.Product{}

			err := data.FromJSON(prod, r.Body)
			if err != nil {
				p.l.Error("error deserealizing product", zap.Error(err))
				http.Error(w, "error reading product", http.StatusBadRequest)
				return
			}

			//validate the product
			errs := p.v.Validate(prod)
			if len(errs) != 0 {
				p.l.Error("[ERROR]", zap.Any("validating product", errs))

				// return the validation messages as an array
				w.WriteHeader(http.StatusUnprocessableEntity)
				data.ToJSON(&ValidationError{Messages: errs.Errors()}, w)
				return
			} else {
				err = data.ToJSON(prod, w)
				if err != nil {
					p.l.Error("[ERROR] error writing success message", zap.Error(err))
					http.Error(w, fmt.Sprintf("writing success message: %s", err), http.StatusBadRequest)
					return
				}
			}

			// add the product to the context
			ctx := InjectProduct(r.Context(), prod)
			ctx = InjectLogger(ctx, p.l)
			r = r.WithContext(ctx)

			p.l.Info("[DEBUG] from middleware", zap.Any("request Info: ", loggableRequest(r)))

			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
}
