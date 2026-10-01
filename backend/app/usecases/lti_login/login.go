// Copyright (c) 2021 MacEwan University. All rights reserved.
//
// This source code is licensed under the MIT-style license found in
// the LICENSE file in the root directory of this source tree.

// Package lti_login provides functions and methods for LTI's modified OpenID Connect login flow.
package lti_login

import (
	"errors"
	"net/url"

	fiber "github.com/gofiber/fiber/v2"

	datastore "github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
	"github.com/google/uuid"
)

const (
	StateCookieName       = "stateCookie"
	LegacyStateCookieName = StateCookieName + "-legacy"
)

// New creates a new login object
func New(cfg *datastore.Config) *Login {
	login := Login{
		cfg: cfg,
	}
	return &login
}

// A Login implements an fiber.Handler that can be easily associated with a tool URI such as /services/lti/login/.
type Login struct {
	cfg *datastore.Config
}

// RedirectURI extracts the form data from the initial login request and returns a auth redirect URI and state cookie.
// The login must cache the "nonce" locally and include it in the response.
func (l *Login) RedirectURI(c *fiber.Ctx) (string, fiber.Cookie, error) {
	registration, err := l.validate(c)
	if err != nil {
		return "", fiber.Cookie{}, err
	}

	// Generate state and state cookie.
	state := "state-" + uuid.New().String()
	stateCookie := fiber.Cookie{
		Name:  StateCookieName,
		Value: state,
		Path:  registration.TargetLinkURI.EscapedPath(),
		// Recent versions of Chrome have changed the default handling of Cookies. To support these versions of
		// Chrome, the following options are necessary.
		//
		// Ref: https://blog.chromium.org/2019/10/developers-get-ready-for-new.html
		SameSite: fiber.CookieSameSiteNoneMode,
		Secure:   true,
	}

	// Generate and store nonce.
	nonce := uuid.New().String()
	err = l.cfg.Nonces.StoreNonce(nonce, registration.TargetLinkURI.String())
	if err != nil {
		return "", fiber.Cookie{}, err
	}

	// Build auth response to initial login request.
	values := url.Values{}
	values.Set("scope", "openid")
	values.Set("response_type", "id_token")
	values.Set("response_mode", "form_post")
	values.Set("prompt", "none")
	values.Set("client_id", registration.ClientID)
	values.Set("redirect_uri", registration.TargetLinkURI.String())
	values.Set("state", state)
	values.Set("nonce", nonce)
	values.Set("login_hint", c.FormValue("login_hint"))

	// Pass back the message hint if received.
	if c.FormValue("lti_message_hint") != "" {
		values.Set("lti_message_hint", c.FormValue("lti_message_hint"))
	}

	redirectURI := registration.AuthLoginURI
	redirectURI.RawQuery = values.Encode()
	return redirectURI.String(), stateCookie, nil
}

// JSRedirect will return JS code to perform the redirect.
func (l *Login) JSRedirect(c *fiber.Ctx) (string, error) {
	redirect, stateCookie, err := l.RedirectURI(c)
	if err != nil {
		return "", err
	}

	return "JS redirect code using: " + redirect + stateCookie.Name, nil
}

// ServeHTTP makes Login an fiber.Ctx so that it can easily be associated with tool URI, e.g., /services/lti/login/.
// The handler must set the "state" in a cookie (in addition to including it in the response) and the two will be
// compared in the launch.
func (l *Login) ServeHTTP(c *fiber.Ctx) error {
	redirectURI, stateCookie, err := l.RedirectURI(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	c.Cookie(&stateCookie)

	if stateCookie.SameSite == fiber.CookieSameSiteNoneMode {
		// Not all browsers support the SameSite=None setting. Create and attach a copy of the cookie without the
		// SameSite=None for these browsers.
		//
		// Ref: https://www.imsglobal.org/samesite-cookie-issues-lti-tool-providers
		legacyStateCookie := stateCookie
		legacyStateCookie.Name = LegacyStateCookieName
		legacyStateCookie.SameSite = fiber.CookieSameSiteDisabled

		c.Cookie(&legacyStateCookie)
	}
	return c.Redirect(redirectURI, fiber.StatusFound)
}

// validate checks for the presence of the issuer and login_hint, and existence of a registration for that issuer.
func (l *Login) validate(c *fiber.Ctx) (datastore.Registration, error) {
	// Validate issuer.
	if c.FormValue("iss") == "" {
		return datastore.Registration{}, errors.New("issuer not found in login request")
	}

	// Validate login hint.
	if c.FormValue("login_hint") == "" {
		return datastore.Registration{}, errors.New("login hint not found in login request")
	}

	// Validate target link uri.
	if c.FormValue("target_link_uri") == "" {
		return datastore.Registration{}, errors.New("target link uri not found in login request")
	}

	// Find Registration by issuer and/or client ID.
	registration, err := l.cfg.Registrations.FindRegistrationByIssuerAndClientID(
		c.FormValue("iss"),
		c.FormValue("client_id"),
	)
	if err != nil {
		return datastore.Registration{}, err
	}

	return registration, nil
}
