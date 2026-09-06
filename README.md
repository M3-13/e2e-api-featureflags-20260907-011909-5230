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

Der Server bindet danach standardmäßig an `127.0.0.1:8080`.

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

## Transport / TLS

Der Dienst bindet standardmäßig an `127.0.0.1:8080` und darf **produktiv
ausschließlich hinter einem TLS-terminierenden Reverse-Proxy** betrieben werden.
Der direkte unverschlüsselte Betrieb ist **nicht** der dokumentierte Standard:
Die `user`-ID aus `GET /flags/{key}/evaluate?user={id}` wird sonst im Klartext
über das Netz übertragen. Der Reverse-Proxy beendet TLS und reicht den Verkehr
intern an den Dienst weiter.

## API-Key

Alle `/flags`-Routen verlangen den Header `X-API-Key`, dessen Wert über die
Umgebungsvariable `API_KEY` gesetzt wird. Ohne gültigen Key antworten sie mit
`401`.

```sh
API_KEY=<geheimer-key> go run .
```

```sh
curl -H "X-API-Key: <geheimer-key>" http://127.0.0.1:8080/flags
```

## Datenschutz / Betrieb

- Die `user`-ID wird ausschließlich **transient** für den FNV-Hash der
  Flag-Auswertung verwendet. Sie wird **nicht persistiert** und **nicht in
  App-Logs geschrieben** (geloggt werden nur Methode, Pfad ohne Query-String
  und Statuscode).
- Der vorgelagerte Reverse-Proxy darf den **Query-String nicht loggen**,
  damit die `user`-ID nicht in Proxy-Logs landet.
- Rechtsgrundlage und datenschutzrechtliche Verantwortlichkeit für die
  Verarbeitung der `user`-ID legt der **Betreiber** fest und bildet sie in
  seinen eigenen Datenschutzhinweisen ab.

## Security / SBOM / Updates

### Abhängigkeiten (SBOM)

Der Dienst verwendet ausschließlich die Go-Standardbibliothek; es gibt **keine
externen Module**. Die vollständige Abhängigkeitsliste ergibt sich aus:

```sh
go list -m all
```

bzw. aus `go.mod` (kein `require`-Block, nur `module featureflags` und
`go 1.22`).

### Versions- und Patch-Prozess

- Sicherheitsrelevante Änderungen werden als neue Version (Patch/Minor) über
  den regulären Release-Prozess ausgeliefert.
- Der Dienst wird vor dem Ausrollen mit `go build ./...` gebaut und mit
  `go test ./...` geprüft.
- Abhängigkeiten (aktuell keine externen Module) werden beim Hinzufügen auf
  bekannte Schwachstellen geprüft und mit einer expliziten Versionsgrenze
  gebunden.

### Meldeweg für Sicherheitslücken

Sicherheitslücken bitte vertraulich an den Sicherheitskontakt des Betreibers
melden (Kontakt und Rechtsgrundlage legt der Betreiber fest). Details zu
Schwachstellen werden erst nach Bereitstellung einer Korrektur veröffentlicht.
