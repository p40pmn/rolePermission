package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/phuangpheth/rolePermission/database"
)

var (
	// ErrUnknownUser is returned when the user is not found.
	ErrUnknownUser = errors.New("unknown user")
)

// Config is the configuration for the middleware.
type Config struct {
	DB *database.DB

	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper
}

// User represents the user information.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

// Role represents the role.
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

// Permission represents the permission.
type Permission struct {
	ID       string `json:"id"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

// Role is the middleware for role.
func RoleMiddleware(cfg Config) echo.MiddlewareFunc {
	if cfg.Skipper == nil {
		cfg.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			user, err := cfg.getUserByID(c.Request().Context(), c.Request().Header.Get("userId"))
			if err == ErrUnknownUser {
				return &echo.HTTPError{
					Code:    http.StatusUnauthorized,
					Message: "unauthorized",
				}
			}
			if err != nil {
				return &echo.HTTPError{
					Code:     http.StatusInternalServerError,
					Message:  "internal server error",
					Internal: errors.New("internal server error"),
				}
			}

			perms, err := cfg.getPermissionByUserID(c.Request().Context(), user.ID)
			if err != nil {
				return &echo.HTTPError{
					Code:     http.StatusInternalServerError,
					Message:  "internal server error",
					Internal: errors.New("internal server error"),
				}
			}

			permsMap := permissionToMap(perms)
			// check permission, if the permission is not found, return error.
			id := fmt.Sprintf("%s-%s", methodToAction(c.Request().Method), strings.Split(c.Path(), "/")[2])
			if _, ok := permsMap[id]; !ok {
				return &echo.HTTPError{
					Code:    http.StatusForbidden,
					Message: "access denied",
				}
			}
			user.Role.Permissions = perms

			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, userKey, user)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

type ctxKey int

const userKey ctxKey = iota

// ClaimUserFromContext returns the user from the context.
func ClaimUserFromContext(ctx context.Context) *User {
	claim, ok := ctx.Value(userKey).(*User)
	if !ok {
		return &User{}
	}
	return claim
}

func (r *Config) getUserByID(ctx context.Context, id string) (User, error) {
	query, args, err := sq.Select("u.id", "u.name", "u.email", "r.id", "r.name").
		From("users AS u").
		JoinClause("INNER JOIN roles AS r ON r.id = u.role_id").
		Where(sq.Eq{"u.id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return User{}, err
	}

	var user User
	row := r.DB.QueryRow(ctx, query, args...)
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.Role.ID, &user.Role.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUnknownUser
	}
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *Config) getPermissionByUserID(ctx context.Context, userID string) ([]Permission, error) {
	query, args, err := sq.Select("p.id", "p.action", "p.resource").
		From("users AS u").
		JoinClause("INNER JOIN roles AS r ON r.id = u.role_id").
		JoinClause("JOIN role_policies AS rp ON rp.role_id = r.id").
		JoinClause("JOIN permissions AS p ON p.id = rp.permission_id").
		Where(sq.Eq{"u.id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	perms := make([]Permission, 0)
	collection := func(rows *sql.Rows) error {
		p, err := scanPermission(rows.Scan)
		if err != nil {
			return err
		}
		perms = append(perms, p)
		return nil
	}
	return perms, r.DB.RunQueryIncrementally(ctx, query, 100, collection, args...)
}

func scanPermission(scan func(...any) error) (p Permission, _ error) {
	return p, scan(&p.ID, &p.Action, &p.Resource)
}

func permissionToMap(perms []Permission) map[string]Permission {
	m := make(map[string]Permission, 0)
	for _, p := range perms {
		id := fmt.Sprintf("%s-%s", p.Action, p.Resource)
		m[id] = p
	}
	return m
}

func methodToAction(method string) string {
	switch method {
	case http.MethodGet:
		return "read"
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	}
	return ""
}
