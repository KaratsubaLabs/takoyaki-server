package main

import (
	"time"
    "net/http"
	"encoding/json"
)

type ContextKey string
func (c ContextKey) String() string {
	return "takoyaki:contextKey:" + string(c)
}

var (
	ContextKeyUserID     = ContextKey("userid")
	ContextKeyParsedBody = ContextKey("parsedbody")
)

type CustomHandler = func(http.ResponseWriter, *http.Request) error

type routeInfo struct {
	route        string
	methods      []string // possibly restrict to certain strings (ie POST, GET)
	authRoute    bool
	bodySchema   interface{}
	handlerFn    CustomHandler
}

func (info routeInfo) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var handlerWithMiddleware http.Handler = ErrorMiddleware(info.handlerFn)

    // validate + parse body (if applicable)
    if info.bodySchema != nil {
		handlerWithMiddleware = ValidationMiddleware(handlerWithMiddleware)
        handlerWithMiddleware = ParseBodyJSONMiddleware(info.bodySchema, handlerWithMiddleware)
    }

	// restrict auth (if applicable)
	if info.authRoute {
		handlerWithMiddleware = AuthMiddleware(handlerWithMiddleware)
	}

	// restrict method of request
	handlerWithMiddleware = RestrictMethodMiddleware(info.methods, handlerWithMiddleware)

	// delegate to handler
	handlerWithMiddleware.ServeHTTP(w, r)

}

var routeSchema = []routeInfo{
	{
		route: "/ping",
		methods: []string{"POST"},
		authRoute: false,
		handlerFn: pingHandler,
	},
	{
		route: "/register",
		methods: []string{"POST"},
		authRoute: false,
		bodySchema: &registerRequest{},
		handlerFn: registerHandler,
	},
	{
		route: "/login",
		methods: []string{"POST"},
		authRoute: false,
		bodySchema: &loginRequest{},
		handlerFn: loginHandler,
	},
	{
		route: "/vps/info",
		methods: []string{"GET"},
		authRoute: true,
		bodySchema: &vpsInfoRequest{},
		handlerFn: vpsInfoHandler,
	},
	{
		route: "/vps/create",
		methods: []string{"POST"},
		authRoute: true,
		bodySchema: &vpsCreateRequest{},
		handlerFn: vpsCreateHandler,
	},
	{
		route: "/vps/delete",
		methods: []string{"POST"},
		authRoute: true,
		bodySchema: &vpsDeleteRequest{},
		handlerFn: vpsDeleteHandler,
	},
}

// ping endpoint for debug purposes
func pingHandler(w http.ResponseWriter, r *http.Request) error {

	// var newVPS = VPSConfig{
	// 	DisplayName: "pino-vps",
	// 	Hostname:    "pino-vps",
	// 	Username:    "pinosaur",
	// 	Password:    "$6$rounds=4096$Z7a9LgphTzzWHJbQ$Yp8C0xPXMJhE45/Q7JLo/OoAWODjlCDGH/Zdgb7FUaX5HeGdnYH4XXP13bWZldzDlSndSKSmDWTbot88ZRuJJ1",
	// 	SSHKey:      "ssh-rsa blah blah",
	// 	RAM:         RAM_HIGH,
	// 	CPU:         1,
	// 	Disk:        25,
	// 	OS:          "ubuntu",
	// }

	// err := VPSCreate(newVPS)
	// if err != nil {
	// 	return HTTPStatusError{http.StatusInternalServerError, err}
	// }

	db, err := DBConnection()
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

	err = DBMigrate(db)
	if err != nil {
		return HTTPStatusError{http.StatusInternalServerError, err}
	}

	return nil
}

type registerRequest struct {
    Username      string         `json:"username" validate:"required,min=3,max=32"`
    Password      string         `json:"password" validate:"required,min=8,max=128"`
	Email         string         `json:"email"    validate:"required,email"`
}
type registerResponse struct {
    Token         string         `json:"token"`
}
func registerHandler(w http.ResponseWriter, r *http.Request) error {

	parsedBody, ok := r.Context().Value(ContextKeyParsedBody).(*registerRequest)
	if !ok {
        return HTTPStatusError{http.StatusInternalServerError, nil}
	}

	// (possibly have db connection be part of the context)
	db, err := DBConnection()
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

	// make sure user name and email are not already taken

	// maybe encrypt password, could be done on frontend

	newUser := User{
		Username: parsedBody.Username,
		Email:    parsedBody.Email,
		Password: parsedBody.Password,
	}
    userID, err := DBUserRegister(db, newUser)
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

	token, err := GenerateToken(userID)
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

    json.NewEncoder(w).Encode(registerResponse{Token: token})

    return nil
}

type loginRequest struct {
	Username      string         `json:"username" validate:"required"`
	Password      string         `json:"password" validate:"required"`
}
type loginResponse struct {
    Token         string         `json:"token"`
}
func loginHandler(w http.ResponseWriter, r *http.Request) error {

	parsedBody, ok := r.Context().Value(ContextKeyParsedBody).(*loginRequest)
	if !ok {
        return HTTPStatusError{http.StatusInternalServerError, nil}
	}

	db, err := DBConnection()
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

	userID, err := DBUserCheckCreds(db, parsedBody.Username, parsedBody.Password)
	if err != nil {
        return HTTPStatusError{http.StatusUnauthorized, err}
	}

	token, err := GenerateToken(userID)
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

    json.NewEncoder(w).Encode(loginResponse{Token: token})

	return nil
}

type vpsInfoRequest struct {
	VPSName      string          `json:"vps_name" validate:"required"`
}
func vpsInfoHandler(w http.ResponseWriter, r *http.Request) error {

	userID, ok := r.Context().Value(ContextKeyUserID).(uint)
	if !ok {
        return HTTPStatusError{http.StatusInternalServerError, nil}
	}

	db, err := DBConnection()
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

	allUserVPS, err := DBVPSGetInfo(db, userID)
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

	_ = allUserVPS

	// TODO figure out how to store user vps config in db

	return nil
}

// or just use the VPSConfig struct directly
type vpsCreateRequest struct {
	DisplayName   string         `json:"display_name" validate:"required,max=128"`
	Hostname      string         `json:"hostname"     validate:"required,max=128"`
	Username      string         `json:"username"     validate:"required,max=32"`
	Password      string         `json:"password"     validate:"required"`
	SSHKey        string         `json:"ssh_key"      validate:""`
	RAM           int            `json:"ram"          validate:"required"`
	CPU           int            `json:"cpu"          validate:"required"`
	Disk          int            `json:"disk"         validate:"required"`
	OS            string         `json:"os"           validate:"required"`
	Message       string         `json:"message"`
}
func vpsCreateHandler(w http.ResponseWriter, r *http.Request) error {

	parsedBody, ok := r.Context().Value(ContextKeyParsedBody).(*vpsCreateRequest)
	if !ok {
        return HTTPStatusError{http.StatusInternalServerError, nil}
	}

	userID, ok := r.Context().Value(ContextKeyUserID).(uint)
	if !ok {
        return HTTPStatusError{http.StatusInternalServerError, nil}
	}

	db, err := DBConnection()
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

	config := VPSCreateRequestData{
		DisplayName: parsedBody.DisplayName,
		Hostname:    parsedBody.Hostname,
		Username:    parsedBody.Username,
		Password:    parsedBody.Password,
		SSHKey:      parsedBody.SSHKey,
		RAM:         parsedBody.RAM,
		CPU:         parsedBody.CPU,
		Disk:        parsedBody.Disk,
		OS:          parsedBody.OS,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

	newRequest := Request{
		UserID:         userID,
		RequestTime:    time.Now(),
		RequestPurpose: REQUEST_PURPOSE_VPS_CREATE,
		RequestData:    string(configJSON),
		Message:        parsedBody.Message,
	}
	err = DBRequestCreate(db, newRequest)
	if err != nil {
        return HTTPStatusError{http.StatusInternalServerError, err}
	}

	return nil
}

type vpsDeleteRequest struct {
	VPSName      string          `json:"vps_name" validate:"required"`
}
func vpsDeleteHandler(w http.ResponseWriter, r *http.Request) error {

	// no need to make this a request - just go ahead and delete

	// issue delete commands

	// remove from db

	return nil
}

func Routes(mux *http.ServeMux) {
	for _, routeInfo := range routeSchema {
		mux.Handle(routeInfo.route, routeInfo)
	}
}

