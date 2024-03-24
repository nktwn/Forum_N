package handlers

import (
	"errors"
	"forum/models"
	"forum/pkg/cookie"
	"forum/pkg/validator"
	"net/http"
)

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	methodResolver(w, r, h.loginGet, h.loginPost)
}

func (h *handler) loginGet(w http.ResponseWriter, r *http.Request) {
	data, err := h.NewTemplateData(r)
	if err != nil {
		h.app.ServerError(w, err)
	}
	data.Form = models.UserLoginForm{}
	h.app.Render(w, http.StatusOK, "login.html", data)
}

func (h *handler) loginPost(w http.ResponseWriter, r *http.Request) {
	form := models.UserLoginForm{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	session, err := h.service.Authenticate(form.Email, form.Password)

	if !form.Valid() {
		data, err := h.NewTemplateData(r)
		if err != nil {
			h.app.ServerError(w, err)
		}
		data.Form = form
		h.app.Render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			form.AddFieldError("email", "email doesn't exist")
			data, err := h.NewTemplateData(r)
			if err != nil {
				h.app.ServerError(w, err)
			}
			data.Form = form
			h.app.Render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddFieldError("password", models.ErrInvalidCredentials.Error())
			data, err := h.NewTemplateData(r)
			if err != nil {
				h.app.ServerError(w, err)
			}
			data.Form = form
			h.app.Render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			h.app.ServerError(w, err)
		}
		return
	}
	cookie.SetSessionCookie(w, session.Token, session.ExpTime)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) signup(w http.ResponseWriter, r *http.Request) {
	methodResolver(w, r, h.signupGet, h.signupPost)
}

func (h *handler) signupGet(w http.ResponseWriter, r *http.Request) {
	data, err := h.NewTemplateData(r)
	if err != nil {
		h.app.ServerError(w, err)
	}
	data.Form = models.UserSignupForm{}
	h.app.Render(w, http.StatusOK, "signup.html", data)
}

func (h *handler) signupPost(w http.ResponseWriter, r *http.Request) {
	form := models.UserSignupForm{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data, err := h.NewTemplateData(r)
		if err != nil {
			h.app.ServerError(w, err)
		}
		data.Form = form
		h.app.Render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}
	//
	user := form.FormToUser()
	err := h.service.CreateUser(user)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data, err := h.NewTemplateData(r)
			if err != nil {
				h.app.ServerError(w, err)
			}
			data.Form = form
			h.app.Render(w, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			h.app.ServerError(w, err)
		}
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *handler) logoutPost(w http.ResponseWriter, r *http.Request) {
	c := cookie.GetSessionCookie(r)
	if c != nil {
		h.service.DeleteSession(c.Value)
		cookie.ExpireSessionCookie(w)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *handler) PostByUser(w http.ResponseWriter, r *http.Request) {
	c := cookie.GetSessionCookie(r)
	posts, err := h.service.GetAllPostByUser(c.Value)
	if err != nil {
		h.app.ServerError(w, err)
		return
	}

	data, err := h.NewTemplateData(r)
	if err != nil {
		h.app.ServerError(w, err)
	}

	data.Categories, err = h.service.GetAllCategory()
	if err != nil {
		h.app.ServerError(w, err)
	}

	data.Posts = posts

	token := cookie.GetSessionCookie(r)
	if token != nil {
		reactions, err := h.service.GetReactionPosts(token.Value)
		if err != nil {
			h.app.ServerError(w, err)
			return
		}
		data.Posts = h.service.IsLikedPost(data.Posts, reactions)
	}

	h.app.Render(w, http.StatusOK, "home.html", data)
}

func (h *handler) LikedPosts(w http.ResponseWriter, r *http.Request) {
	c := cookie.GetSessionCookie(r)
	posts, err := h.service.GetLikedPosts(c.Value)
	if err != nil {
		h.app.ServerError(w, err)
		return
	}

	data, err := h.NewTemplateData(r)
	if err != nil {
		h.app.ServerError(w, err)
	}

	data.Categories, err = h.service.GetAllCategory()
	if err != nil {
		h.app.ServerError(w, err)
	}

	data.Posts = posts

	token := cookie.GetSessionCookie(r)
	if token != nil {
		reactions, err := h.service.GetReactionPosts(token.Value)
		if err != nil {
			h.app.ServerError(w, err)
			return
		}
		data.Posts = h.service.IsLikedPost(data.Posts, reactions)
	}

	h.app.Render(w, http.StatusOK, "home.html", data)
}
