# chirpy

A simple HTTP server built with Go for creating and storing chirps: Small message blurbs to describe your thoughts, opinions, etc...

## API Endpoints Overview

Method | Path | Description | Auth Required | Request Body | Notes
| --- | --- | --- | --- | --- | --- |
POST | /api/users | Create a new user | No | Email, Password | Signup endpoint
POST | /api/login | Login and get tokens | No | Email, Password | Returns access + refresh tokens
POST | /api/refresh | Refresh access token | Yes (refresh token) | None | Uses refresh token
POST | /api/revoke | Revoke refresh token | Yes (refresh token) | None | Logout by invalidating refresh token
POST | /api/reset | Reset database (dev only) | No | None | Only works in dev mode
GET | /api/metrics | Get file server hit metrics | No | None | Returns hit count
GET | /api/chirps | Get all chirps | No | None | Supports sort and author_id query params
GET | /api/chirps/{chirpID} | Get a chirp by ID | No | None | 404 if not found
DELETE | /api/chirps/{chirpID} | Delete a chirp | Yes (access token) | None | Only owner can delete
PUT | /api/users | Update user's email/password | Yes (access token) | Email and/or Password | Partial updates allowed
POST | /api/upgrade | Upgrade user to Chirpy Red | Yes (Polka API key) | Event payload | Called from external API

- Auth Required:
    - "No": Public Endpoint
    - "Yes": Means you must pass a token in ```Authorization: Bearer <token>```.


## Authentication Guide

Some endpoints require authentication. Here's how to authenticate:
1. Access Token (Bearer Token)

Used for most user-authenticated actions.

Header Example:

Authorization: Bearer <access_token>

    Get your access token by logging in (POST /api/login).

    Include it in the Authorization header for protected endpoints like deleting chirps or updating your profile.

2. Refresh Token

Used to refresh your access token (POST /api/refresh) or revoke it (POST /api/revoke).

Header Example:

Authorization: Bearer <refresh_token>

    Treat refresh tokens like access tokens but only use them for refresh/revoke endpoints.

    They usually have longer expiration times.

3. Polka API Key

Used by external services to trigger upgrades (like Chirpy Red subscriptions).

Header Example:

Authorization: ApiKey <your_polka_api_key>

    Only the /api/upgrade endpoint expects this format.

    This is not a user token — it's a special secret given to trusted third parties.
