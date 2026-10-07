# API contract

The `.proto` files define every RPC between the frontend and the backend. This
file covers the parts protobuf can't express.

## Transport

[Connect](https://connectrpc.com) over HTTP, with `connect-go` on the backend
and `@connectrpc/connect-web` on the frontend. Mount every Connect handler under
`/api`, so a method lives at `/api/autovolume.v1.AuthService/StartLogin`:

```go
api := http.NewServeMux()
api.Handle(autovolumev1connect.NewAuthServiceHandler(authSvc))
// ...
mux.Handle("/api/", http.StripPrefix("/api", api))
```

## Plain HTTP routes

| Route | Purpose |
| --- | --- |
| `GET /api/img/{rating_key}?w=&h=` | Poster proxy. Fetches the Plex `/photo/:/transcode` with the server token and sets cache headers. `MediaItem.thumb_url` points here. |
| Everything else | The built frontend, falling back to `index.html` for client-side routes like `/login/callback`. |

The backend learns about playback from the Plex server's notification
websocket, not webhooks, so Plex never calls in.

## Sessions

The session is an HttpOnly cookie set by `AuthService.CompleteLogin`. The
frontend sends it on every request and never handles a Plex token.

## Error codes

The frontend routes on these codes, so keep them consistent:

| Code | Meaning to the frontend |
| --- | --- |
| `Unauthenticated` | No valid session: go to login. Use it only for the app session, never for TV or Plex auth failures. |
| `FailedPrecondition` | Setup is incomplete (no server, player or TV): go to setup. |
| `PermissionDenied` | Shown inline, e.g. a wrong TV PSK, or a Plex account without access to the selected server. |
| `Unavailable` | Plex server or TV unreachable. Shown inline. |
| `InvalidArgument` | Bad input, e.g. a volume outside 0-100. Shown inline. |

Put a human-readable message on errors; the UI displays it.

## Codegen

From the repo root, run:

```sh
buf lint
buf generate
```

This writes Go to `backend/gen/` (import path `autovolume/gen/autovolume/v1`)
and TypeScript to `web/src/gen/`. Both directories are wiped and regenerated
each time, so don't edit them.
