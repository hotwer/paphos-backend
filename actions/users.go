package actions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"

	"paphos/models"
	"paphos/shared"
)

func UsersShowGet(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(200, r.JSON(user))
}

// Registers a new User.
func UsersRegisterPost(c buffalo.Context) error {
	// Allocate an empty User
	user := &models.User{}

	// Bind user to the request body payload
	if err := c.Bind(user); err != nil {
		return err
	}

	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Validate the data from the request
	verrs, err := user.Create(tx)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusCreated, r.JSON(user))
}

// Logs a user in.
func UsersLoginPost(c buffalo.Context) error {
	user := &models.User{}

	if err := c.Bind(user); err != nil {
		return err
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	err, populated_user := user.Authorize(tx)
	if err != nil {
		verrs := validate.NewErrors()
		verrs.Add("root", "Invalid email or password.")
		return c.Error(http.StatusForbidden, verrs)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":           populated_user.ID,
		"email":        populated_user.Email,
		"display_name": populated_user.DisplayName,

		"iat": jwt.NewNumericDate(time.Now()),
		"exp": jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	})

	tokenString, err := token.SignedString(shared.JWT_SECRET)
	if err != nil {
		return err
	}

	return c.Render(http.StatusCreated, r.JSON(struct {
		ID          uuid.UUID `json:"id"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		JWT         string    `json:"jwt"`
	}{populated_user.ID, populated_user.Email, populated_user.DisplayName, tokenString}))
}