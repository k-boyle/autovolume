# autovolume

Sets your Sony Bravia TV's or Sony soundbar's volume based on what Plex is
playing.

Some films are mixed quiet and some shows are mixed loud. autovolume lets you
give a movie, show or season its own volume. When playback starts on your
chosen Plex player, it sets the TV to that volume. When playback pauses or
stops, it puts the TV back to your default.

## How it works

The backend connects to your Plex Media Server's notification websocket and
watches playback on one player. On each state change it picks a volume and
sends it to the device you hear sound from, over Sony's REST API:

| Playback | Volume |
| --- | --- |
| Movie playing | The movie's rule, else the default |
| Episode playing | The season's rule, else the show's rule, else the default |
| Paused or stopped | The default |

It only acts when the state changes. Plex re-sends "playing" every ~10 seconds,
and re-applying the volume each time would override the TV remote. Until you
set a default volume, autovolume leaves the TV alone when nothing has a rule.

## Requirements

- A Plex Media Server owned by your Plex account. Only owners receive playback
  notifications.
- One of:
  - a Sony Bravia TV with IP control enabled and a pre-shared key set, or
  - a Sony soundbar that supports the Audio Control API. It needs no key.
- PostgreSQL.
- Network access from autovolume to both the Plex server and the TV or
  soundbar.

### TV setup

On the TV, open **Settings → Network & Internet → Home network → IP control**.
Set **Authentication** to **Normal and Pre-Shared Key**, then set a
**Pre-Shared Key**. Menu names vary a little between models.

In autovolume, enter the address as `http://<tv-ip>`, along with the key.

### Soundbar setup

If your sound goes through a Sony soundbar, point autovolume at the soundbar
instead of the TV. The TV's own volume does nothing while the soundbar is
playing.

Enter the address as `http://<soundbar-ip>:10000` and leave the key blank.

## Running with Docker

Images are published to Docker Hub as `kieranthebest/autovolume` on every push
to `main`.

```yaml
services:
  autovolume:
    image: kieranthebest/autovolume:latest
    ports:
      - "8080:8080"
    environment:
      AV_DATABASE_CONNECTION: postgres://autovolume:autovolume@db:5432/autovolume?sslmode=disable
    depends_on:
      - db

  db:
    image: postgres:18
    environment:
      POSTGRES_USER: autovolume
      POSTGRES_PASSWORD: autovolume
      POSTGRES_DB: autovolume
    volumes:
      - pgdata:/var/lib/postgresql

volumes:
  pgdata:
```

Migrations run automatically on startup. Then open `http://localhost:8080`:

1. Sign in with Plex.
2. Pick your server.
3. Pick the player to watch. A player may only appear once it has played
   something.
4. Enter the TV or soundbar address (and the TV's pre-shared key), and test
   the connection.

From the home page you can browse your libraries and set per-item volumes. The
default volume lives in Settings.

## Configuration

Settings are read from `config.json` in the working directory, and can be
overridden with environment variables prefixed `AV_`.

| Key | Env var | Default | Description |
| --- | --- | --- | --- |
| `port` | `AV_PORT` | `8080` | HTTP listen port |
| `database_connection` | `AV_DATABASE_CONNECTION` | none | PostgreSQL connection string |

Everything else (Plex credentials, server, player, TV, volumes) is stored in
the database and set through the web UI.

## Development

The repo has three parts:

| Path | What it is |
| --- | --- |
| `backend/` | Go server. Serves the API under `/api` and embeds the built frontend. |
| `web/` | React + Vite + Tailwind frontend. |
| `proto/` | The API contract. See [`proto/README.md`](proto/README.md). |

The frontend and backend talk over [Connect](https://connectrpc.com).

### Backend

The Go binary embeds `backend/web/dist`, so build the frontend once before
running it:

```sh
npm --prefix web ci
npm --prefix web run build
AV_DATABASE_CONNECTION=postgres://... go -C backend run .
```

### Frontend

```sh
npm --prefix web run dev
```

The Vite dev server proxies `/api` to `http://localhost:8080`. Set
`AUTOVOLUME_BACKEND` to point it somewhere else.

To work on the UI without a backend, Plex server or TV, run it against mock
data:

```sh
npm --prefix web run dev:mock
```

Lint with `npm --prefix web run lint`.

### Regenerating the API

After changing a `.proto` file, run this from the repo root (needs
[buf](https://buf.build) and `web/node_modules` installed):

```sh
buf lint
buf generate
```

This rewrites `backend/gen/` and `web/src/gen/`. Don't edit those by hand.

### Docker image

```sh
docker build -t autovolume .
```

## License

[MIT](LICENSE)
