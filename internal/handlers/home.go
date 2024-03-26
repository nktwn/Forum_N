package handlers

import (
	"errors"
	"forum/models"
	"forum/pkg/cookie"
	"net/http"
	"strconv"
)

func (h *handler) home(w http.ResponseWriter, r *http.Request) {
	data, err := h.NewTemplateData(r)
	if err != nil {
		h.app.ServerError(w, err)
	}
	data, err = h.service.SetUpPage(data, r)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			h.app.NotFound(w)
			return
		} else {
			h.app.ServerError(w, err)
			return
		}
	}
	if data.Category_id == 0 {
		posts, err := h.service.GetAllPostPaginated(data.CurrentPage, data.Limit)
		if err != nil {
			h.app.ServerError(w, err)
			return
		}

		data.Posts = posts
	} else {
		posts, err := h.service.GetAllPostByCategoryPaginated(data.CurrentPage, data.Limit, data.Category_id)
		if err != nil {
			h.app.ServerError(w, err)
			return
		}
		data.Posts = posts
	}
	token := cookie.GetSessionCookie(r)
	if token != nil {
		reactions, err := h.service.GetReactionPosts(token.Value)
		if err != nil {
			h.app.ServerError(w, err)
			return
		}
		data.Posts = h.service.IsLikedPost(data.Posts, reactions)
	}

	if len(*data.Posts) == 0 {
		data.Posts = nil
	}

	h.app.Render(w, http.StatusOK, "home.html", data)
	return
}

// func (h *handler) homePost(w http.ResponseWriter, r *http.Request) {
// 	if err := r.ParseForm(); err != nil {
// 		h.app.ClientError(w, http.StatusBadRequest)
// 		return
// 	}

// 	filterCategoriesString := r.Form["categories"]
// 	filterCategories, err := ConverCategories(filterCategoriesString)
// 	if err != nil {
// 		h.app.ClientError(w, http.StatusBadRequest)
// 	}
// 	posts, err := h.service.GetAllPostByCategories(filterCategories)
// 	if err != nil {
// 		h.app.ServerError(w, err)
// 	}
// 	data := h.NewTemplateData(r)

// 	data.Categories, err = h.service.GetAllCategory()
// 	if err != nil {
// 		h.app.ServerError(w, err)
// 	}

// 	data.Posts = posts
// 	h.app.Render(w, http.StatusOK, "home.html", data)
// }

func ConverCategories(CategoriesString []string) ([]int, error) {
	categories := make([]int, len(CategoriesString))
	for i, str := range CategoriesString {
		nb, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		categories[i] = nb
	}

	return categories, nil
}
