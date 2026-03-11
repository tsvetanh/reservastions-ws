package utils

import "net/http"

const (
	// Error Codes
	SUCCESS             = 0
	INVALID_REQ_PAYLOAD = iota + 1000
	FAILED_HASH_PASSWORD
	MISSING_TOKEN
	INVALID_TOKEN_FORMAT
	INVALID_TOKEN
	INVALID_TOKEN_CLAIMS
	USERNAME_NOT_IN_TOKEN
	USER_NOT_IN_CONTEXT
	USER_NOT_FOUND
	INVALID_USER_ID_PARAM
	ACCESS_DENIED
	FAILED_GET_USERS
	FAILED_UPDATE_ROLE
	FAILED_ASSIGN_ROLE
	FAILED_REVOKE_ROLE
	FAILED_GET_ROLES
	ROLE_NOT_FOUND
	INVALID_ROLE_ID_PARAM
	MISSING_IMAGES
	USERNAME_EXISTS
	FAILED_CREATE_ROLE
	FAILED_CREATE_USER
	FAILED_CREATE_TOKEN
	INVALID_CREDENTIALS
	TOKEN_CREATION_FAILED
	INVALID_HALL_DATA
	MISSING_HALL_DATA
	INVALID_DATE_RANGE
	CAPACITY_AND_COST_ERROR
	AVAILABLE_FROM_IN_PAST
	INVALID_RES_START_DATE
	FAILED_CREATE_HALL
	FAILED_SAVE_IMAGE
	FAILED_SAVE_IMAGE_DATA
	IMAGE_NOT_FOUND
	FAILED_GET_HALLS
	FAILED_UPDATE_HALL
	FAILED_DELETE_HALL
	HALL_NOT_FOUND
	INVALID_HALL_ID
	INVALID_DATE_FORMAT
	FAILED_GET_RESERVATIONS
	HALL_BOOKED
	FAILED_CREATE_RESERVATION
	FAILED_DELETE_RESERVATION
	FAILED_CREATE_RECEIPT
	FAILED_DELETE_RECEIPT
	RESERVATION_NOT_FOUND
	FAILED_UPDATE_RESERVATION
	NO_RESERVATIONS
)

type ErrorDetails struct {
	HTTPStatus int
	Message    string
}

var ErrorMessages = map[int]ErrorDetails{
	SUCCESS: {
		HTTPStatus: http.StatusOK,
		Message:    "Success",
	},
	INVALID_REQ_PAYLOAD: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "Invalid request payload",
	},
	FAILED_HASH_PASSWORD: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to hash password",
	},
	MISSING_TOKEN: {
		HTTPStatus: http.StatusUnauthorized,
		Message:    "Missing token",
	},
	INVALID_TOKEN_FORMAT: {
		HTTPStatus: http.StatusUnauthorized,
		Message:    "Invalid token format",
	},
	INVALID_TOKEN: {
		HTTPStatus: http.StatusUnauthorized,
		Message:    "Invalid or expired token",
	},
	INVALID_TOKEN_CLAIMS: {
		HTTPStatus: http.StatusUnauthorized,
		Message:    "Invalid token claims",
	},
	USERNAME_NOT_IN_TOKEN: {
		HTTPStatus: http.StatusUnauthorized,
		Message:    "Username not found in token claims",
	},
	USER_NOT_IN_CONTEXT: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "'user' not found in gin context",
	},
	USER_NOT_FOUND: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "User not found",
	},
	INVALID_USER_ID_PARAM: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "User id should be a number",
	},
	ACCESS_DENIED: {
		HTTPStatus: http.StatusForbidden,
		Message:    "Access denied",
	},
	FAILED_GET_USERS: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to retrieve users",
	},
	FAILED_UPDATE_ROLE: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to update role",
	},
	FAILED_ASSIGN_ROLE: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to assign role",
	},
	FAILED_REVOKE_ROLE: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to revoke role",
	},
	FAILED_GET_ROLES: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to retrieve roles",
	},
	ROLE_NOT_FOUND: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Role not found",
	},
	INVALID_ROLE_ID_PARAM: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "Role id should be a number",
	},
	MISSING_IMAGES: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "No images in form-data found",
	},
	USERNAME_EXISTS: {
		HTTPStatus: http.StatusConflict,
		Message:    "Username already exists",
	},
	FAILED_CREATE_ROLE: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to create role",
	},
	FAILED_CREATE_USER: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to create user",
	},
	FAILED_CREATE_TOKEN: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to create token",
	},
	INVALID_CREDENTIALS: {
		HTTPStatus: http.StatusUnauthorized,
		Message:    "Invalid credentials",
	},
	TOKEN_CREATION_FAILED: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Could not create token",
	},
	INVALID_HALL_DATA: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "Invalid hall data provided",
	},
	MISSING_HALL_DATA: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "Hall data missing",
	},
	INVALID_DATE_RANGE: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "AvailableFrom must be before AvailableTo",
	},
	CAPACITY_AND_COST_ERROR: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "Capacity and cost must be positive numbers",
	},
	AVAILABLE_FROM_IN_PAST: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "AvailableFrom cannot be in the past",
	},
	INVALID_RES_START_DATE: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "Start date must be between today and end date",
	},
	FAILED_CREATE_HALL: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to create hall",
	},
	FAILED_GET_HALLS: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to retrieve halls",
	},
	FAILED_DELETE_HALL: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to delete hall",
	},
	FAILED_UPDATE_HALL: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to update hall",
	},
	FAILED_SAVE_IMAGE: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to save image",
	},
	FAILED_SAVE_IMAGE_DATA: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to save image data",
	},
	IMAGE_NOT_FOUND: {
		HTTPStatus: http.StatusNotFound,
		Message:    "Image not found",
	},
	INVALID_HALL_ID: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "Invalid hall ID",
	},
	INVALID_DATE_FORMAT: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "Invalid date format",
	},
	FAILED_GET_RESERVATIONS: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to retrieve reservations",
	},
	HALL_BOOKED: {
		HTTPStatus: http.StatusConflict,
		Message:    "Hall is already booked for these dates",
	},
	FAILED_CREATE_RESERVATION: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to create reservation",
	},
	FAILED_DELETE_RESERVATION: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to delete reservation\"",
	},
	FAILED_CREATE_RECEIPT: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Reservation created, but failed to generate receipt",
	},
	FAILED_DELETE_RECEIPT: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Reservation deleted, but failed to delete receipt",
	},
	RESERVATION_NOT_FOUND: {
		HTTPStatus: http.StatusNotFound,
		Message:    "Reservation not found",
	},
	FAILED_UPDATE_RESERVATION: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Failed to update reservation",
	},
	NO_RESERVATIONS: {
		HTTPStatus: http.StatusNoContent,
		Message:    "No reservations found",
	},
}
