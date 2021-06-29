package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"

	"github.com/storj/dms/pkg/storage"
)

// DMSRouter contains all the routes for the DMS functionality
type DMSRouter struct {
	Store               storage.StorageImpl
	JWTSecret           string
	HeartbeatExpiration time.Duration
}

// Pingdom is the endpoint that pingdom will query, it returns 200 if all the checks are within the heartbearExpiration
// and it returns 500 if any of the checks are outside of the heartbearExpiration
func (r *DMSRouter) Pingdom(c echo.Context) error {
	all, errStore := r.Store.All()
	if errStore != nil {
		logrus.WithError(errStore).Error("unable to retrieve all stored data from etcd")
	}

	var envsNotChecked []string
	for k, v := range all {
		if v.Last != (time.Time{}) {
			if time.Since(v.Last) > (r.HeartbeatExpiration) {
				logrus.Debugf("environment=%s has not checked in for %s", k, r.HeartbeatExpiration)
				envsNotChecked = append(envsNotChecked, k)
			}
		}
	}

	var errMsgs []string
	if errStore != nil {
		errMsgs = append(errMsgs, errStore.Error())
	}

	if len(envsNotChecked) > 0 {
		errMsgs = append(errMsgs, fmt.Sprintf("the following services have not checked in: %s", strings.Join(envsNotChecked, ", ")))
	}

	if len(errMsgs) > 0 {
		return c.JSON(http.StatusInternalServerError, strings.Join(errMsgs, ". "))
	}

	return c.JSON(http.StatusOK, "all services have checked in")
}

// Register generates a JWT for use in authenticating against the /ingest route
func (r *DMSRouter) Register(c echo.Context) error {
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["environment"] = c.FormValue("environment")

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(r.JWTSecret))
	if err != nil {
		return err
	}

	r.Store.Store(claims["environment"].(string), storage.Data{})

	return c.JSON(http.StatusCreated, map[string]string{
		"token": t,
	})
}

// Status is an internal endpoint
func (r *DMSRouter) Status(c echo.Context) error {
	all, err := r.Store.All()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, all)
}

// Ingest is the primary heartbeat route
func (r *DMSRouter) Ingest(c echo.Context) error {
	source := c.Get("user").(*jwt.Token)
	claims := source.Claims.(jwt.MapClaims)
	environment := claims["environment"].(string)

	r.Store.Store(environment, storage.Data{
		Last: time.Now(),
	})

	return c.String(http.StatusOK, "from "+environment)
}
