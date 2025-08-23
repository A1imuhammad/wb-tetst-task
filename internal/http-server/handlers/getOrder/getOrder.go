package getorder

import (
	"demoserv/internal/cache"
	"demoserv/internal/postgress"
	
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, cache *cache.Cache, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		order_uid := chi.URLParam(r, "order_uid")

		// Получаем из кэша
		if order, ok := cache.Get(order_uid); ok {
			render.JSON(w, r, order)
			return
		}

		// Получаем из бд если нет в кэше
		order, err := postgress.GetOrder(ctx, order_uid, pool)
		if err != nil {
			log.Printf("order %s not found in db", order_uid)
			http.Error(w, fmt.Sprintf(`{"error":"order %s not found"}`, order_uid), http.StatusNotFound)
			return
		}

		// Добавляем в кэш после получения
		cache.Add(order)
		// Отправляем ответ
		render.JSON(w, r, order)

	}

}



