# Feature-Flag-Service

Ein in Go geschriebener Feature-Flag-Service als REST-API. Flags werden in einem
thread-sicheren In-Memory-Store verwaltet; die Auswertung liefert über einen
stabilen Hash deterministische Ja/Nein-Entscheidungen pro Nutzer.

## Tech-Stack

- **language**: Go (1.22)
- **framework**: `net/http` aus der Standardbibliothek (kein externes Web-Framework)
- **build**: `go.mod`
- **testing**: `go test`, `httptest`

## Installation

```sh
go mod download
```

## Starten (Dev)

```sh
go run .
```

Der Server lauscht danach auf `http://localhost:8080`.

## Build (Production)

```sh
go build ./...
```

## Verwendung

### Endpunkte

| Methode | Pfad                        | Beschreibung                                             |
| ------- | --------------------------- | -------------------------------------------------------- |
| POST    | `/flags`                    | Flag anlegen (`{key, enabled, description?, rollout_percent?}`) |
| GET     | `/flags`                    | Alle Flags als JSON-Array                                |
| GET     | `/flags/{key}`              | Einzelnes Flag                                            |
| PUT     | `/flags/{key}`              | Flag aktualisieren                                        |
| DELETE  | `/flags/{key}`              | Flag löschen (204)                                        |
| GET     | `/flags/{key}/evaluate?user={id}` | Deterministische Auswertung (`{"result":bool}`)     |
| GET     | `/healthz`                  | Health-Check (`{"status":"ok"}`)                          |

Fehlerantworten tragen immer die Form `{"error":"..."}`.

## Feature-Liste

- Thread-sicherer In-Memory-Store (`sync.RWMutex`)
- Anlegen, Auflisten, Lesen, Aktualisieren und Löschen von Flags
- Deterministische Rollout-Auswertung pro Nutzer über einen stabilen Hash
- Health-Endpoint `/healthz`
