VERDICT: BLOCKED

## Sicherheitsbericht

**Scanner-Analyse:** Für den Projekttyp `go-backend` wurden laut Vorgabe keine anwendbaren Security-Scanner ausgeführt (`no applicable security scanners for this project type`). Die folgende Bewertung basiert ausschließlich auf der manuellen Analyse des sichtbaren Codes.

---

### Kritisch/Hoch

#### 1. Fehlende Authentifizierung und Autorisierung für sämtliche API-Endpunkte
- **Schweregrad:** hoch
- **Betroffen:** `main.go:23-32` (Serverstart und Routing), `handlers_create.go`, `handlers_update.go`, `handlers_delete.go`, `handlers_get.go`, `handlers_list.go`, `handlers_evaluate.go`
- **Beschreibung:** Der Server lauscht mit `http.ListenAndServe(":8080", handler)` auf allen Netzwerkinterfaces und stellt sämtliche Endpunkte ohne jede Authentifizierung oder Autorisierung bereit. Jeder, der den Port erreichen kann, darf:
  - Feature Flags anlegen (`POST /flags`)
  - bestehende Flags verändern (`PUT /flags/{key}`)
  - Flags löschen (`DELETE /flags/{key}`)
  - Flags und Rollout-Entscheidungen auslesen (`GET /flags`, `GET /flags/{key}`, `GET /flags/{key}/evaluate`)
- **Risiko:** Ein Angreifer kann unautorisierte Feature-Aktivierungen erzwingen, den Dienst durch Löschen aller Flags lahmlegen oder vertrauliche Produktentscheidungen auslesen. Das ist ein klassischer „auth bypass“ durch fehlende Zugriffskontrolle.
- **Fix:** Vor den mutierenden und lesenden Routen eine Authentifizierungs-/Autorisierungs-Middleware einfügen, z. B. einen API-Key oder ein Bearer-Token prüfen. Zusätzlich sollte der Dienst nicht ungeschützt auf `:8080` an allen Interfaces lauschen. Konkrete Maßnahmen:
  - Middleware `RequireAuth(next http.Handler)` implementieren, die `Authorization`- oder `X-API-Key`-Header gegen ein konfigurierbares Geheimnis prüft.
  - Alternativ den Server initial an `127.0.0.1:8080` binden und die Authentifizierung über einen vorgeschalteten Reverse-Proxy erzwingen.
  - Die Auth-Anforderung ist mit der Produktfunktion vereinbar: Clients müssen lediglich den konfigurierten Schlüssel mitsenden; die API selbst bleibt funktionsfähig.

---

### Mittel

#### 2. `rollout_percent` wird nicht auf den gültigen Bereich 0–100 validiert
- **Schweregrad:** mittel
- **Betroffen:** `handlers_create.go:23-29`, `handlers_update.go:24-29`, `store_create.go:7-16`, `store_update.go:6-18`, `store_evaluate.go:17`
- **Beschreibung:** Beim Anlegen und Aktualisieren wird `rollout_percent` als beliebiger `int` akzeptiert. Die Auswertung in `store_evaluate.go` verwendet:
  ```go
  return int(h.Sum32()%100) < flag.RolloutPercent, nil
  ```
  Ein Wert > 100 (z. B. 101) führt dazu, dass das Flag für **alle** Nutzer aktiv ist; ein negativer Wert führt dazu, dass es für **keinen** Nutzer aktiv ist. Das ermöglicht ungewollte Feature-Rollouts und widerspricht der erwarteten Prozent-Semantik.
- **Fix:** Vor dem Speichern in `CreateHandler` und `UpdateHandler` prüfen:
  ```go
  if body.RolloutPercent < 0 || body.RolloutPercent > 100 {
      writeJSONError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
      return
  }
  ```
  Alternativ oder zusätzlich im Store (`Create`, `Update`) validieren und einen Fehler zurückgeben.

---

### Niedrig

#### 3. Recover-Middleware kann Panic-Details und potenziell User-IDs ins Log schreiben
- **Schweregrad:** niedrig
- **Betroffen:** `middleware.go:37`
- **Beschreibung:** Die Recover-Middleware verwendet:
  ```go
  log.Printf("panic recovered: %v", err)
  ```
  Der Panic-Wert kann interne Details enthalten – im schlimmsten Fall auch die in einem Handler verarbeitete User-ID aus `GET /flags/{key}/evaluate`. Das Prinzip der Datensparsamkeit in Logs (AC-17) bezieht sich zwar primär auf die Logging-Middleware, aber auch die Recover-Middleware schreibt ungefilterte Fehlerwerte.
- **Fix:** Protokollieren Sie nur eine generische Meldung:
  ```go
  log.Println("panic recovered")
  ```
  Detaillierte Fehlerwerte oder Stacktraces nur in einem separaten, internen Debug-Kanal mit Zugriffsbeschränkung ablegen.

#### 4. Kein TLS auf dem HTTP-Endpunkt
- **Schweregrad:** niedrig
- **Betroffen:** `main.go:31`
- **Beschreibung:** Der Server lauscht unverschlüsselt auf `:8080`. Sowohl Flag-Inhalte als auch der Query-Parameter `user` werden im Klartext übertragen. Je nach Netzwerkposition können Dritte diese Daten mitlesen.
- **Fix:** TLS-Terminierung durch einen Reverse-Proxy (empfohlen) oder direkte Verwendung von `http.ListenAndServeTLS`. Mindestens sollte der Dienst an ein privates Interface gebunden werden.

#### 5. Format des Flag-`key` nicht beschränkt
- **Schweregrad:** niedrig
- **Betroffen:** `handlers_create.go:23-29`, `store_create.go:7-16`
- **Beschreibung:** Der Schlüssel wird ohne Längen- oder Zeichenbeschränkung übernommen. Schlüssel mit `/` können über die Pfad-Routen `GET/PUT/DELETE /flags/{key}` nicht sauber adressiert werden, da `{key}` genau ein Pfadsegment umfasst. Sonderzeichen können außerdem die Lesbarkeit der JSON-Ausgabe beeinträchtigen.
- **Fix:** Vor dem Anlegen einen regulären Ausdruck wie `^[A-Za-z0-9._-]+$` und eine Maximallänge erzwingen; bei Verstoß 400 zurückgeben.

---

**Fazit:** Die spezifizierten Einzelsicherheitsmaßnahmen (Body-Limit, Content-Type-Prüfung, keine CORS-Header, Logging ohne Query-String, Recover ohne Stacktrace im Body) sind weitgehend korrekt umgesetzt. Der schwerwiegende Mangel ist die vollständig fehlende Authentifizierung/Autorisierung für eine mutierende HTTP-API, die auf allen Interfaces lauscht. Deshalb wird das Produkt derzeit als **BLOCKED** eingestuft, bis der Zugriffsschutz nachgerüstet ist.