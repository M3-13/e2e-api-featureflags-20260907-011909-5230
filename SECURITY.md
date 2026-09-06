VERDICT: CHANGES_REQUESTED

## Sicherheitsbericht

### 1. Zusammenfassung der Prüfbereiche

- **Secrets:** Keine hartcodierten Geheimnisse gefunden. Der API-Key wird aus der Umgebungsvariable `API_KEY` gelesen.  
- **Injection & Inputs:** Key-Validierung verhindert Path-Traversal; JSON-Bodies werden größenbegrenzt; keine SQL/Command-/Deserialisierungs-Endpunkte.  
- **AuthN/AuthZ:** API-Key-Authentifizierung mit `crypto/subtle.ConstantTimeCompare`; nur `GET /healthz` ist ausgenommen. Kein Auth-Bypass erkennbar.  
- **Dependencies:** `go.mod` enthält ausschließlich die Standardbibliothek. Es liegen keine Scanner-Ergebnisse vor (keine anwendbaren Scanner für dieses Projekt).  
- **Configuration/Transport:** Server bindet auf `127.0.0.1:8080`; Timeouts gesetzt; keine CORS-Header; TLS nicht erforderlich, da nur localhost.  

Das Produkt erfüllt die spezifizierten Sicherheitsanforderungen weitgehend. Die folgenden Findings sind Härtungsempfehlungen und ein mittelgradiger DoS-Aspekt durch die Middleware-Reihenfolge.

---

### 2. Findings

#### F1 – Mittel: Unauthentifizierter Ressourcenverbrauch durch Middleware-Reihenfolge
- **Betroffen:** `main.go` – `handler := Recover(Logging(BodyLimit(ContentType(RequireAuth(mux)))))`  
- **Beschreibung:** `BodyLimit` und `ContentType` werden **vor** `RequireAuth` ausgeführt. Ein Angreifer ohne gültigen API-Key kann dadurch große POST/PUT-Requests (bis 1 MB) senden, die von `BodyLimit` vollständig eingelesen und im Speicher gehalten werden, **bevor** die Authentifizierung greift. Dies ist ein unnötiger DoS-Vektor und liefert außerdem unterschiedliche Statuscodes (z. B. 415 bei falschem Content-Type, 413 bei zu großem Body) statt einheitlich 401, was einem Angreifer Hinweise auf die API-Struktur geben kann.  
- **Empfehlung/Fix:** Authentifizierung vor die ressourcenintensiven Middleware-Schichten setzen, z. B.  
  ```go
  handler := Recover(Logging(RequireAuth(BodyLimit(ContentType(mux)))))
  ```
  Sicherstellen, dass `RequireAuth` für `GET /healthz` weiterhin ohne Key durchlässt. Die Funktionalität der Produkt-Routen bleibt unverändert.

---

#### F2 – Niedrig: Content-Type-Prüfung ist case-sensitive
- **Betroffen:** `middleware.go` – `mediaType(ct string)` / `ContentType`  
- **Beschreibung:** HTTP-Media-Types sind laut RFC case-insensitive. Der aktuelle Vergleich `mediaType(ct) != "application/json"` akzeptiert nur exakt kleingeschriebene Werte. Ein Client, der `Application/JSON` sendet, wird mit 415 abgewiesen, obwohl der Header gültig ist. Dies ist primär ein funktionales Problem, kann aber in Kombination mit F1 als zusätzliches Informationsleck dienen.  
- **Empfehlung/Fix:** Den Media-Type vor dem Vergleich normalisieren:  
  ```go
  func mediaType(ct string) string {
      if i := strings.IndexByte(ct, ';'); i >= 0 {
          ct = ct[:i]
      }
      return strings.ToLower(strings.TrimSpace(ct))
  }
  ```
  Dann mit `strings.EqualFold` oder dem normalisierten Wert vergleichen.

---

#### F3 – Niedrig: `Recover` nach bereits gesendetem Header
- **Betroffen:** `middleware.go` – `Recover` und `writeJSONError`  
- **Beschreibung:** Falls ein Handler bereits `WriteHeader` aufgerufen hat und anschließend eine Panic auslöst (z. B. während des JSON-Encodings), versucht `Recover`, erneut `writeJSONError` zu senden. Da die Header bereits geschrieben wurden, schlägt der Aufruf fehl; der Client erhält eine unvollständige oder inkonsistente Antwort. Der Log-Status (`statusRecorder`) kann zudem den ursprünglich gesetzten Status (z. B. 200) anzeigen, nicht 500.  
- **Empfehlung/Fix:** In `Recover` prüfen, ob bereits geschrieben wurde. Da `statusRecorder` den Status erfasst, kann ein eigenes `Recover`-Wrapper-Flag verwendet werden. Alternativ Handler so strukturieren, dass `WriteHeader` erst nach vollständiger Verarbeitung erfolgt. Beispielhafter Ansatz:  
  ```go
  func Recover(next http.Handler) http.Handler {
      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          defer func() {
              if err := recover(); err != nil {
                  log.Print("panic recovered")
                  if !isHeaderWritten(w) { // eigene Implementierung oder ResponseWriter-Wrapper
                      writeJSONError(w, http.StatusInternalServerError, "internal server error")
                  }
              }
          }()
          next.ServeHTTP(w, r)
      })
  }
  ```
  Einfacher ist, alle Handler so zu bauen, dass sie erst nach vollständiger Antwortberechnung `WriteHeader` aufrufen – dies ist in den bestehenden Handlern der Fall. Das Finding ist daher vorbeugend.

---

#### F4 – Niedrig: Kein Rate-Limiting für API-Key-Brute-Force
- **Betroffen:** `auth.go` – `RequireAuth`  
- **Beschreibung:** Obwohl `crypto/subtle.ConstantTimeCompare` den Vergleich zeitkonstant macht, gibt es keinen Schutz gegen fortgesetzte Fehlversuche. Da der Dienst auf `127.0.0.1` lauscht, ist das Risiko lokal begrenzt. Sollte der Dienst jedoch hinter einem Reverse-Proxy erreichbar gemacht werden, wäre ein Brute-Force-Angriff auf den API-Key möglich.  
- **Empfehlung/Fix:** Optional einen einfachen Rate-Limiter pro Client-IP oder eine exponentielle Verzögerung bei wiederholten 401-Antworten einführen. Für den aktuellen localhost-Betrieb nicht zwingend.

---

### 3. Positiv hervorgehoben

- **Key-Validierung:** `^[A-Za-z0-9._-]+$` und Längenlimit 128 verhindern Path-Traversal und Sonderzeichen.
- **Body-Limit:** Größenbegrenzung auf 1 MB mit `http.MaxBytesReader` und zusätzlicher `ContentLength`-Prüfung funktioniert wie spezifiziert.
- **Logging:** Protokolliert ausschließlich Methode, `r.URL.Path` (ohne Query) und Statuscode; keine Benutzer-IDs oder Query-Parameter.
- **Panic-Behandlung:** `Recover` gibt ein generisches JSON-Fehlerobjekt zurück und loggt keine Panic-Details.
- **Keine CORS-Header:** Es werden keine `Access-Control-Allow-*`-Header gesetzt.
- **Transport:** Bindung an `127.0.0.1:8080`, Timeouts für Read/Write/Header/Idle sind vorhanden.

### 4. Scanner-Lücke

Im bereitgestellten Scanner-Output sind für dieses Projekt (`go-backend`) keine anwendbaren Security-Scanner aufgeführt. Das Fehlen von Scanner-Ergebnissen wurde nicht als Befund gewertet; die manuelle Code-Analyse deckt die sichtbaren Schwachstellen ab.