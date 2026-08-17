package docs

import "github.com/swaggo/swag"

const docTemplate = `{
  "schemes": {{ marshal .Schemes }},
  "swagger": "2.0",
  "info": {
    "description": "Backend API for EasyTranslator: games, translations, auth, notifications, moderation, chat, and scanner callbacks.",
    "title": "EasyTranslator API",
    "contact": {},
    "version": "1.0"
  },
  "host": "{{ .Host }}",
  "basePath": "{{ .BasePath }}",
  "paths": {
    "/": {
      "get": {
        "summary": "Health check",
        "tags": ["System"],
        "responses": { "200": { "description": "API is running" } }
      }
    },
    "/api/auth/register": {
      "post": {
        "summary": "Register user",
        "tags": ["Auth"],
        "consumes": ["application/json"],
        "produces": ["application/json"],
        "parameters": [{ "in": "body", "name": "body", "required": true, "schema": { "$ref": "#/definitions/RegisterRequest" } }],
        "responses": { "201": { "description": "Created" }, "400": { "description": "Bad request" } }
      }
    },
    "/api/auth/login": {
      "post": {
        "summary": "Login user",
        "tags": ["Auth"],
        "consumes": ["application/json"],
        "produces": ["application/json"],
        "parameters": [{ "in": "body", "name": "body", "required": true, "schema": { "$ref": "#/definitions/LoginRequest" } }],
        "responses": { "200": { "description": "JWT token" }, "401": { "description": "Unauthorized" } }
      }
    },
    "/health": { "get": { "summary": "Health check", "tags": ["System"], "responses": { "200": { "description": "OK" } } } },
    "/api/auth/me": {
      "get": {
        "summary": "Get current user",
        "tags": ["Auth"],
        "security": [{ "BearerAuth": [] }],
        "responses": { "200": { "description": "Current user" }, "401": { "description": "Unauthorized" } }
      }
    },
    "/cards": {
      "get": {
        "summary": "List game cards",
        "tags": ["Games"],
        "produces": ["application/json"],
        "responses": { "202": { "description": "Game cards", "schema": { "type": "array", "items": { "$ref": "#/definitions/GameCard" } } } }
      }
    },
    "/games": {
      "get": {
        "summary": "List games with translations",
        "tags": ["Games"],
        "produces": ["application/json"],
        "responses": { "202": { "description": "Games", "schema": { "type": "array", "items": { "$ref": "#/definitions/PublicGameInfo" } } } }
      }
    },
    "/games/{gameid}": {
      "get": {
        "summary": "Get game by id",
        "tags": ["Games"],
        "produces": ["application/json"],
        "parameters": [{ "name": "gameid", "in": "path", "required": true, "type": "integer" }],
        "responses": { "202": { "description": "Game", "schema": { "$ref": "#/definitions/PublicGameInfo" } }, "400": { "description": "Bad request" } }
      },
      "delete": {
        "summary": "Delete game",
        "tags": ["Games"],
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "gameid", "in": "path", "required": true, "type": "integer" }],
        "responses": { "200": { "description": "Deleted" }, "401": { "description": "Unauthorized" }, "403": { "description": "Forbidden" } }
      }
    },
    "/games/add": {
      "post": {
        "summary": "Create game",
        "tags": ["Games"],
        "security": [{ "BearerAuth": [] }],
        "consumes": ["multipart/form-data"],
        "parameters": [
          { "name": "Title", "in": "formData", "required": true, "type": "string" },
          { "name": "big_pic", "in": "formData", "required": true, "type": "file" },
          { "name": "small_pic", "in": "formData", "required": true, "type": "file" }
        ],
        "responses": { "201": { "description": "Created" }, "400": { "description": "Bad request" } }
      }
    },
    "/games/translate/{gameid}": {
      "post": {
        "summary": "Upload translation archive",
        "tags": ["Translations"],
        "security": [{ "BearerAuth": [] }],
        "consumes": ["multipart/form-data"],
        "parameters": [
          { "name": "gameid", "in": "path", "required": true, "type": "integer" },
          { "name": "file", "in": "formData", "required": true, "type": "file" },
          { "name": "authorName", "in": "formData", "type": "string" },
          { "name": "source", "in": "formData", "type": "string" },
          { "name": "version", "in": "formData", "type": "number", "format": "double" },
          { "name": "percentReady", "in": "formData", "type": "number", "format": "double" }
        ],
        "responses": { "201": { "description": "Created and queued for scan" }, "400": { "description": "Bad request" } }
      }
    },
    "/games/translate/{transid}": {
      "delete": {
        "summary": "Delete translation",
        "tags": ["Translations"],
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "transid", "in": "path", "required": true, "type": "integer" }],
        "responses": { "200": { "description": "Deleted" }, "404": { "description": "Not found" } }
      }
    },
    "/download/{transid}": {
      "get": {
        "summary": "Download translation archive",
        "tags": ["Translations"],
        "parameters": [
          { "name": "transid", "in": "path", "required": true, "type": "integer" },
          { "name": "token", "in": "query", "type": "string", "description": "Optional JWT token for browser downloads" }
        ],
        "responses": { "200": { "description": "File" }, "403": { "description": "Forbidden" }, "404": { "description": "Not found" } }
      }
    },
    "/api/files/hash-check": { "post": { "summary": "Check archive SHA-256 duplicate", "tags": ["Translations"], "security": [{ "BearerAuth": [] }], "parameters": [{ "in": "body", "name": "body", "required": true, "schema": { "$ref": "#/definitions/ArchiveHashCheckRequest" } }], "responses": { "200": { "description": "Exists flag" }, "400": { "description": "Bad request" } } } },
    "/api/me/translations": { "get": { "summary": "List current user translations", "tags": ["Translations"], "security": [{ "BearerAuth": [] }], "responses": { "200": { "description": "Translations", "schema": { "type": "array", "items": { "$ref": "#/definitions/MyTranslation" } } } } } },
    "/translations/{transid}/status": { "get": { "summary": "Get translation scan status", "tags": ["Translations"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "transid", "in": "path", "required": true, "type": "integer" }], "responses": { "200": { "description": "Status", "schema": { "$ref": "#/definitions/TranslationStatus" } }, "403": { "description": "Forbidden" }, "404": { "description": "Not found" } } } },
    "/translations/{transid}/files": { "get": { "summary": "List approved translation files", "tags": ["Translations"], "parameters": [{ "name": "transid", "in": "path", "required": true, "type": "integer" }], "responses": { "200": { "description": "Files", "schema": { "type": "array", "items": { "$ref": "#/definitions/PublicGameFile" } } }, "403": { "description": "Forbidden" }, "404": { "description": "Not found" } } } },
    "/translations/{transid}": { "delete": { "summary": "Delete own pending/rejected translation", "tags": ["Translations"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "transid", "in": "path", "required": true, "type": "integer" }], "responses": { "200": { "description": "Deleted" }, "403": { "description": "Forbidden" }, "404": { "description": "Not found" } } } },
    "/api/admin/users": {
      "get": {
        "summary": "List users",
        "tags": ["Admin"],
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "page", "in": "query", "type": "integer" },
          { "name": "limit", "in": "query", "type": "integer" }
        ],
        "responses": { "200": { "description": "Users" }, "403": { "description": "Forbidden" } }
      }
    },
    "/api/admin/users/{userid}/block": { "patch": { "summary": "Block user", "tags": ["Admin"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "userid", "in": "path", "required": true, "type": "integer" }], "responses": { "200": { "description": "OK" } } } },
    "/api/admin/users/{userid}/unblock": { "patch": { "summary": "Unblock user", "tags": ["Admin"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "userid", "in": "path", "required": true, "type": "integer" }], "responses": { "200": { "description": "OK" } } } },
    "/api/admin/users/{userid}/warn": { "patch": { "summary": "Warn user", "tags": ["Admin"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "userid", "in": "path", "required": true, "type": "integer" }, { "in": "body", "name": "body", "schema": { "$ref": "#/definitions/ReasonRequest" } }], "responses": { "200": { "description": "OK" } } } },
    "/api/admin/users/{userid}/unwarn": { "patch": { "summary": "Remove user warning", "tags": ["Admin"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "userid", "in": "path", "required": true, "type": "integer" }], "responses": { "200": { "description": "OK" } } } },
    "/api/admin/users/{userid}/role": { "patch": { "summary": "Set user role", "tags": ["Admin"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "userid", "in": "path", "required": true, "type": "integer" }, { "in": "body", "name": "body", "schema": { "$ref": "#/definitions/RoleRequest" } }], "responses": { "200": { "description": "OK" } } } },
    "/api/notifications": { "get": { "summary": "Get my notifications", "tags": ["Notifications"], "security": [{ "BearerAuth": [] }], "responses": { "200": { "description": "Notifications" } } } },
    "/api/admin/notifications": { "post": { "summary": "Create notification", "tags": ["Notifications"], "security": [{ "BearerAuth": [] }], "parameters": [{ "in": "body", "name": "body", "required": true, "schema": { "$ref": "#/definitions/NotificationRequest" } }], "responses": { "201": { "description": "Created" } } } },
    "/api/admin/moderation": { "get": { "summary": "Get moderation queue", "tags": ["Moderation"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "page", "in": "query", "type": "integer" }, { "name": "limit", "in": "query", "type": "integer" }], "responses": { "200": { "description": "Queue" } } } },
    "/api/admin/moderation/{transid}/approve": { "patch": { "summary": "Approve translation", "tags": ["Moderation"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "transid", "in": "path", "required": true, "type": "integer" }], "responses": { "200": { "description": "OK" } } } },
    "/api/admin/moderation/{transid}/reject": { "patch": { "summary": "Reject translation", "tags": ["Moderation"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "transid", "in": "path", "required": true, "type": "integer" }, { "in": "body", "name": "body", "schema": { "$ref": "#/definitions/ReasonRequest" } }], "responses": { "200": { "description": "OK" } } } },
    "/api/admin/moderation/{transid}/change-status/{status}": { "patch": { "summary": "Change translation status", "tags": ["Moderation"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "transid", "in": "path", "required": true, "type": "integer" }, { "name": "status", "in": "path", "required": true, "type": "string" }], "responses": { "200": { "description": "OK" } } } },
    "/api/chat/history/{userId}": { "get": { "summary": "Get chat history", "tags": ["Chat"], "security": [{ "BearerAuth": [] }], "parameters": [{ "name": "userId", "in": "path", "required": true, "type": "integer" }], "responses": { "200": { "description": "Messages" } } } },
    "/api/chat/ws": { "get": { "summary": "Chat WebSocket", "tags": ["Chat"], "security": [{ "BearerAuth": [] }], "responses": { "101": { "description": "Switching Protocols" } } } },
    "/api/internal/scan-result": { "post": { "summary": "Scanner callback", "tags": ["Internal"], "security": [{ "InternalKeyAuth": [] }], "parameters": [{ "in": "body", "name": "body", "required": true, "schema": { "$ref": "#/definitions/ScanResultRequest" } }], "responses": { "200": { "description": "Accepted" }, "400": { "description": "Bad request" }, "403": { "description": "Forbidden" } } } }
  },
  "securityDefinitions": {
    "BearerAuth": { "type": "apiKey", "name": "Authorization", "in": "header", "description": "JWT token in format: Bearer <token>" },
    "InternalKeyAuth": { "type": "apiKey", "name": "X-Internal-Key", "in": "header" }
  },
  "definitions": {
    "RegisterRequest": { "type": "object", "properties": { "firstName": { "type": "string" }, "lastName": { "type": "string" }, "nickname": { "type": "string" }, "password": { "type": "string" } } },
    "LoginRequest": { "type": "object", "properties": { "nickname": { "type": "string" }, "password": { "type": "string" } } },
    "ReasonRequest": { "type": "object", "properties": { "reason": { "type": "string" } } },
    "RoleRequest": { "type": "object", "properties": { "role": { "type": "string", "example": "moderation" } } },
    "NotificationRequest": { "type": "object", "properties": { "title": { "type": "string" }, "message": { "type": "string" }, "isGlobal": { "type": "boolean" }, "userId": { "type": "integer" } } },
    "ArchiveHashCheckRequest": { "type": "object", "properties": { "archiveHash": { "type": "string" } } },
    "PublicGameInfo": { "type": "object", "properties": { "id": { "type": "integer" }, "title": { "type": "string" }, "iconUrl": { "type": "string" }, "translations": { "type": "array", "items": { "$ref": "#/definitions/PublicTranslationSummary" } } } },
    "PublicTranslationSummary": { "type": "object", "properties": { "id": { "type": "integer" }, "authorName": { "type": "string" }, "source": { "type": "string" }, "version": { "type": "number" }, "percentReady": { "type": "number" }, "fileSize": { "type": "number" }, "createdAt": { "type": "string", "format": "date-time" }, "downloadUrl": { "type": "string" } } },
    "PublicGameFile": { "type": "object", "properties": { "fileName": { "type": "string" }, "size": { "type": "string" } } },
    "TranslationStatus": { "type": "object", "properties": { "id": { "type": "integer" }, "status": { "type": "string" }, "scanDetails": { "type": "string" }, "files": { "type": "array", "items": { "$ref": "#/definitions/PublicGameFile" } } } },
    "MyTranslation": { "type": "object", "properties": { "id": { "type": "integer" }, "gameInfoId": { "type": "integer" }, "authorName": { "type": "string" }, "source": { "type": "string" }, "version": { "type": "number" }, "percentReady": { "type": "number" }, "fileSize": { "type": "number" }, "status": { "type": "string" }, "scanDetails": { "type": "string" }, "createdAt": { "type": "string", "format": "date-time" }, "downloadUrl": { "type": "string" } } },
    "GameCard": { "type": "object", "properties": { "id": { "type": "integer" }, "title": { "type": "string" }, "iconUrl": { "type": "string" }, "gameId": { "type": "integer" } } },
    "GameInfo": { "type": "object", "properties": { "id": { "type": "integer" }, "title": { "type": "string" }, "iconUrl": { "type": "string" }, "translateCards": { "type": "array", "items": { "$ref": "#/definitions/TranslateCard" } } } },
    "TranslateCard": { "type": "object", "properties": { "id": { "type": "integer" }, "authorName": { "type": "string" }, "authoreId": { "type": "integer" }, "source": { "type": "string" }, "version": { "type": "number" }, "percentReady": { "type": "number" }, "urlToDownload": { "type": "string" }, "fileSize": { "type": "number" }, "status": { "type": "string" }, "scanDetails": { "type": "string" }, "gameFiles": { "type": "array", "items": { "$ref": "#/definitions/DetailedGameFile" } }, "createdAt": { "type": "string", "format": "date-time" } } },
    "DetailedGameFile": { "type": "object", "properties": { "fileName": { "type": "string" }, "hash": { "type": "string" }, "size": { "type": "string" } } },
    "ScanResultRequest": { "type": "object", "properties": { "transId": { "type": "integer" }, "status": { "type": "string", "example": "approved" }, "details": { "type": "string" }, "threats": { "type": "array", "items": { "type": "string" } }, "error": { "type": "string" }, "files": { "type": "array", "items": { "$ref": "#/definitions/DetailedGameFile" } } } }
  }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "EasyTranslator API",
	Description:      "Backend API for EasyTranslator.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
